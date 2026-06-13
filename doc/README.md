# Go-HAR

A Go library for parsing, manipulating, and generating HTTP Archive (HAR) files.

## Overview

This library provides tools for working with HAR files in Go, offering different parsing strategies optimized for various use cases:

- **Standard parsing**: Simple and straightforward for most use cases
- **Memory-optimized parsing**: Reduced memory footprint for working with large HAR files
- **Lazy loading**: Delayed loading of large content fields until needed
- **Streaming parsing**: Process entries one at a time without loading the entire file

## Installation

```
go get github.com/cyberspacesec/har-skills
```

## Basic Usage

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/cyberspacesec/har-skills"
)

func main() {
    // Parse a HAR file
    harData, err := har.ParseHarFile("example.har")
    if err != nil {
        log.Fatalf("Failed to parse HAR file: %v", err)
    }
    
    // Access HAR data
    fmt.Printf("HAR contains %d entries\n", len(harData.Log.Entries))
    
    // Process entries
    for _, entry := range harData.Log.Entries {
        fmt.Printf("Request: %s %s\n", entry.Request.Method, entry.Request.URL)
        fmt.Printf("Response: %d %s\n", entry.Response.Status, entry.Response.StatusText)
    }
}
```

## Memory Optimization

For large HAR files, use the memory-optimized version:

```go
harData, err := har.ParseHarFileWithOpt(filename, har.ParseOptMemoryOptimized)
```

## Lazy Loading

Lazy loading defers parsing of large fields until needed:

```go
harData, err := har.ParseHarFileWithOpt(filename, har.ParseOptLazyLoad)

// Content is loaded only when explicitly requested
for _, entry := range harData.Log.Entries {
    // Access response content only when needed
    content := entry.Response.Content.Load()
    if content.Size > 1000000 {
        fmt.Printf("Large response found: %s\n", entry.Request.URL)
    }
}
```

## Streaming Processing

Process entries one at a time without loading the entire file:

```go
iterator, err := har.NewStreamingHarParser(filename)
if err != nil {
    log.Fatal(err)
}

for iterator.Next() {
    entry := iterator.Entry()
    // Process each entry individually
    fmt.Printf("URL: %s\n", entry.Request.URL)
}

if err := iterator.Err(); err != nil {
    log.Fatal(err)
}
```

## Advanced Usage

See the examples directory for more advanced usage patterns.