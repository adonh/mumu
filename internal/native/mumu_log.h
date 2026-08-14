#ifndef MUMU_LOG_H
#define MUMU_LOG_H

#import <Foundation/Foundation.h>

// MUMU_LOG prepends a "Mumu: " prefix to all log lines for greppability in
// Console.app and `log show`. Centralized so the prefix can't drift across
// files.
#define MUMU_LOG(fmt, ...) NSLog(@"Mumu: " fmt, ##__VA_ARGS__)

#endif  // MUMU_LOG_H
