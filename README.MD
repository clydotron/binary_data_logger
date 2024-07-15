# Binary Data Logger
Coding homework assignment that implements a thread safe binary data logging service.
Written in Go.


## Components

### Logger

Responsible for reading and writing the binary data to file.
Design details:
- Thread safe
- Uses buffered channel to queue objects to log
- Has internal go routine to:
	- consume loggable object from the channel
	- serialize them, write to storage
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
TODO: replace clazz with LoggableFactory interface.


### Loggable
```
type BinaryLoggable interface {

	// Serialize the fields of this object into a byte array.
	toBytes() ([]byte, error)

	// Deserialize the fields of this object from given byte array.
	fromBytes(rawBytes []byte) error
}
```

### Iterator
Simple iterator interface that returns a generic `interface{}` for each element being iterated over

#### Interface
```
type SimpleIterator interface {
	HasNext() bool
	Next() interface{}
}
```

## Usage



