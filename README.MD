# Binary Data Logger
Small coding project that implements a thread-safe binary data logging service.
Written in Go.


## Components

### Logger

Responsible for reading and writing the binary data to file.
Design details:
- Thread safe
- Uses buffered channel to queue objects to log
- Has internal go routine to:
	- consume loggable object from the channel
	- serialize them
	- write to storage
- Uses bufio.Reader and bufio.Writer for buffered read/write operations

#### Interface
```
type BinaryLogger interface 

	// Write loggable objects to file
	Write(loggable BinaryLoggable) error

	// Read log entries from the specified file via an interator
	// @param clazz instance of the struct to be used for deserializing the data from file.
	Read(file *os.File, clazz any) (SimpleIterator, error)
}
```

### Loggable

An object that can be logged by an object implementing the 'BinaryLogger' iterface. 
Must be able to read/write data to binary format. Each loggable can specify the byte order to use, default is little endian.

```
type BinaryLoggable interface {

	// Serialize the fields of this object into a byte array.
	toBytes() ([]byte, error)

	// Deserialize the fields of this object from given byte array.
	fromBytes(rawBytes []byte) error
}
```

``` 
type BinaryLoggableFactory {
	New() BinaryLoggable
}
```

### Iterator
Simple iterator interface that returns a generic `interface{}` for each element being iterated over.

#### Interface
```
type SimpleIterator interface {
	HasNext() bool
	Next() interface{}
}
```

## Usage
The code demonstrates writing to the logger from multiple concurrent go routines, then retrieving all of the logged data.
You can configure the code by editing these values in the top of the 'main.go' file:


| parameter | value | description |
| --- | --- | --- |
| logFileName | string | name of the file to write/read log messages |
| numLogWriters | int | number of concurrent go routines to run to generate log messages |
| logWriterDurationInSeconds | int | how long (in seconds) to run each log writer |
| maxLogFileSize | int | not currently enforced. Coming soon! |

To run the code: 
'go run .'

### Command line arguments:
Coming Soon


TODO: replace clazz with LoggableFactory interface.


