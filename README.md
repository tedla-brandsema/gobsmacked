# Gobsmacked

**Gobsmacked** is a Go library that extends the functionality of the standard `encoding/gob` package. It introduces additional metadata management, file-level prefixes, and object-level prefixes for serialized GOB data, providing enhanced capabilities like integrity verification, timestamping, and structured storage.

## Features
- **File Metadata**:
    - Format identifier (`gobs`) for distinguishing file types.
    - Versioning support for compatibility across different implementations.
    - Compression and encryption indicators for file-level transformations.
    - Checksum validation for file integrity.
- **Object Metadata**:
    - Object-specific identifiers.
    - Checksums for data integrity.
    - Size prefix to manage object boundaries.
    - Timestamp prefix for object-level event tracking.

## Limitations of GOB Data
The limitations on GOB data size stem from Go's use of the `int` type for functions like `len()` and `cap()`. Since `int` size is architecture-dependent, on 32-bit systems, the maximum allowable size for GOB data is restricted to `math.MaxInt32` bytes (approximately 2 GiB). This library adheres to the 32-bit boundary to maintain compatibility across systems.

To mitigate issues caused by these limitations, Gobsmacked implements structured prefixes that clearly define the size of serialized objects and validate integrity using checksums.

## Installation
Install Gobsmacked using `go get`:

```sh
go get github.com/tedla-brandsema/gobsmacked
```

Import it in your Go project:

```go
import "github.com/tedla-brandsema/gobsmacked"
```

### Adding Metadata
Gobsmacked automatically manages metadata for both files and individual objects. For advanced use cases, you can access or manipulate the metadata prefixes directly using the library's API.

## License
Gobsmacked is released under the MIT License. See [LICENSE](LICENSE) for details.

---

Enhance your Go serialization workflows with **Gobsmacked**! By providing robust metadata and compatibility mechanisms, it makes working with GOB data more manageable and scalable.

