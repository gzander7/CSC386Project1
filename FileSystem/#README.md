FileSys Readme
By Gage Ross
This package implements a simple file system in Go. It has the following features:

Constants:
BLOCK_SIZE: 1024 (1 KB)
DISK_SIZE: 6 * 1024 * 1024 (6 MB)
MAX_DIRECT_BLOCKS: 3
MAX_INODES: 80

Global variables:
VirtualDisk: a 2D array of bytes representing the virtual disk
Inodes: an array of Inode structs representing the inodes
SuperBlock: a struct containing metadata about the file system\

Structs:
Inode: a struct representing an inode, with the following fields:
IsValid: a boolean indicating whether the inode is allocated or not
IsDirectory: a boolean indicating whether the inode represents a directory or a regular file
DataBlocks: an array of four ints representing the data blocks associated with the inode
CreatedTime: a time.Time object representing the time the inode was created
LastModifiedTime: a time.Time object representing the time the inode was last modified

Directory format:
Directories are represented as slices of FileEntry structs, where each FileEntry contains a filename and an inode number.
Filenames are fixed size, with a length of 11 bytes (8 characters plus 3 for the file extension).

Virtual disk layout:
The first block of the virtual disk is reserved for the SuperBlock struct.
Inodes are stored in contiguous blocks starting at block 2.
The allocation bitmap is stored in the remaining blocks of the virtual disk.

Allocation bitmaps:
The data allocation bitmap is implemented as a slice of bools, with one bool per block.
The inode allocation bitmap is implemented as a slice of bools, with one bool per inode.

Functions:
InitializeFileSystem: initializes the file system by setting up the SuperBlock and creating the root directory
ReadSuperBlock: reads the SuperBlock from the disk
EncodeToBytes: encodes a Go object into a byte array
Open: opens a file with the given filename and mode and returns the inode
findInode: finds the inode number of a file with the given filename and parent directory
GetCorrectInode: returns the inode of a file with the given filename and parent directory
Read: reads an inode from the disk
Write: writes an inode to the disk
CreateFile: creates a new file with the given filename and parent directory
findFreeInode: finds a free inode in the inode array
AddFileToDirectory: adds a new file to a directory
ReadInode: reads an inode from the inode array
ReadDataBlock: reads a data block from the disk
WriteDataBlock: writes a data block to the disk
WriteInode: writes an inode to the inode array
Unlink: unlinks a file from its parent directory
WriteFile: writes data to a file
ReadFile: reads data from a file

# Left Undone/Not working
I could not figure out indirection after hours of work a day for about a week straight.
I think that is an area I might need help better understanding in a programming aspect.
