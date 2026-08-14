{
  fetchurl,
  gitUpdater,
  installShellFiles,
  stdenv,
  versionCheckHook,
  lib,
  buildGoModule,
  version ? "main",
  useZip ? false,
  commitHash ? null,
  writableTmpDirAsHomeHook,
  nix-update-script,
  unzip,
  apple-sdk_15,
}:
if useZip then
  let
    appName = "Mumu.app";

    # Determine architecture-specific details
    archInfo =
      {
        "aarch64-darwin" = {
          url = "https://github.com/adonh/mumu/releases/download/v${version}/mumu-darwin-arm64.zip";
          # run `nix hash convert --hash-algo sha256 (nix-prefetch-url https://github.com/adonh/mumu/releases/download/v0.1.0/mumu-darwin-arm64.zip)`
          sha256 = "sha256-3+eQ1htkTXra+n28TuNZ6FwZhgWi2j4DpQ5UybQiIhE=";
        };
        "x86_64-darwin" = {
          url = "https://github.com/adonh/mumu/releases/download/v${version}/mumu-darwin-amd64.zip";
          # run `nix hash convert --hash-algo sha256 (nix-prefetch-url https://github.com/adonh/mumu/releases/download/v0.1.0/mumu-darwin-amd64.zip)`
          sha256 = "sha256-jky8NfaIUTFKftbN0wFBj3VsCJYpDPGHV7BOW+zHLKM=";
        };
      }
      .${stdenv.hostPlatform.system} or (throw "Unsupported system: ${stdenv.hostPlatform.system}");

  in
  stdenv.mkDerivation {
    pname = "mumu";

    inherit version;

    src = fetchurl {
      url = archInfo.url;
      sha256 = archInfo.sha256;
    };

    unpackPhase = ''
      unzip $src
    '';

    nativeBuildInputs = [
      installShellFiles
      unzip
    ];

    installPhase = ''
      runHook preInstall
      ${
        if stdenv.hostPlatform.isDarwin then
          ''
            mkdir -p $out/Applications
            mv ${appName} $out/Applications
            cp -R bin $out
            mkdir -p $out/share/man/man1
            mv share/man/man1/*.1 $out/share/man/man1/
          ''
        else
          ''
            mkdir -p $out/bin
            mv bin/mumu $out/bin/mumu
            mkdir -p $out/share/man/man1
            mv share/man/man1/*.1 $out/share/man/man1/
          ''
      }
      runHook postInstall
    '';

    postInstall = ''
      if ${
        lib.boolToString (
          stdenv.buildPlatform.canExecute stdenv.hostPlatform && stdenv.hostPlatform.isDarwin
        )
      }; then
        installShellCompletion --cmd mumu \
              --bash <($out/Applications/Mumu.app/Contents/MacOS/mumu completion bash) \
              --fish <($out/Applications/Mumu.app/Contents/MacOS/mumu completion fish) \
              --zsh <($out/Applications/Mumu.app/Contents/MacOS/mumu completion zsh)
      fi
    '';

    doInstallCheck = true;
    nativeInstallCheckInputs = [
      versionCheckHook
    ];

    passthru.updateScript = gitUpdater {
      url = "https://github.com/adonh/mumu.git";
      rev-prefix = "v";
    };

    meta = with lib; {
      description = "Save and restore window-to-Space layouts on macOS";
      homepage = "https://github.com/adonh/mumu";
      license = licenses.mit;
      platforms = platforms.darwin;
      mainProgram = "mumu";
    };
  }
else
  let
    shortHash = if commitHash != null then lib.substring 0 7 commitHash else null;

    pversion = "${version}${if shortHash != null then "-${shortHash}" else ""}";
  in
  # Build from source
  buildGoModule (finalAttrs: {
    pname = "mumu";
    version = pversion;

    src = lib.cleanSource ../.;

    # run the following command to get the sha256 hash
    # `nix-shell -p go --run 'go mod vendor'`
    # `nix hash path vendor`
    # `rm -rf vendor`
    vendorHash = "sha256-Ow9rTJNPcpa/pyAwna7Cq4qGt6ph86CpF74q8imwrd4=";

    ldflags = [
      "-s"
      "-w"
      "-X github.com/adonh/mumu/cmd/mumu/cmd.Version=${finalAttrs.version}"
    ]
    ++ lib.optionals (commitHash != null) [
      "-X github.com/adonh/mumu/cmd/mumu/cmd.GitCommit=${commitHash}"
    ];

    subPackages = [ "cmd/mumu" ];

    nativeBuildInputs = [
      installShellFiles
      writableTmpDirAsHomeHook
    ];

    buildInputs = [
      apple-sdk_15
    ];

    # Allow Go to use any available toolchain
    preBuild = ''
      export GOTOOLCHAIN=auto
    '';

    postInstall = ''
      # generate man pages
      mkdir -p $out/share/man/man1
      go run ./cmd/genman $out/share/man/man1

      # install shell completions
      if ${lib.boolToString (stdenv.buildPlatform.canExecute stdenv.hostPlatform)}; then
      	installShellCompletion --cmd mumu \
      	--bash <($out/bin/mumu completion bash) \
      	--fish <($out/bin/mumu completion fish) \
      	--zsh <($out/bin/mumu completion zsh)
      fi
    ''
    + lib.optionalString stdenv.hostPlatform.isDarwin ''
      # Create a simple .app bundle on the fly for macOS source builds.
      mkdir -p $out/Applications/Mumu.app/Contents/{MacOS,Resources}

      cp $out/bin/mumu $out/Applications/Mumu.app/Contents/MacOS/mumu

      # cp ${finalAttrs.src}/resources/icon.icns $out/Applications/Mumu.app/Contents/Resources/icon.icns
      cp ${finalAttrs.src}/resources/Mumu.entitlements $out/Applications/Mumu.app/Contents/Resources/Mumu.entitlements

      SRC_PLIST=${finalAttrs.src}/resources/Info.plist.template

      sed "s|VERSION|${finalAttrs.version}|g" $SRC_PLIST > $out/Applications/Mumu.app/Contents/Info.plist

      echo "✅ Mumu.app bundle created at $out/Applications/Mumu.app"
    '';

    passthru = {
      updateScript = nix-update-script { };
    };

    meta = with lib; {
      description = "Save and restore window-to-Space layouts on macOS";
      homepage = "https://github.com/adonh/mumu";
      license = licenses.mit;
      platforms = platforms.darwin;
      mainProgram = "mumu";
    };
  })
