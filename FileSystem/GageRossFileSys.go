package FileSys

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"
)

const (
	CREATE = iota
	READ
	APPEND
	WRITE
)

const (
	BLOCK_SIZE          = 1024            // 1 KB
	DISK_SIZE           = 6 * 1024 * 1024 // 6 MB
	MAX_DIRECT_BLOCKS   = 3               // Maximum number of direct data blocks
	MAX_INDIRECT_BLOCKS = 1               // Maximum number of indirect data blocks
)

const (
	MAX_INODES = 80 // Maximum number of inodes
)

var VirtualDisk [DISK_SIZE / BLOCK_SIZE][BLOCK_SIZE]byte // Virtual disk represented as a 2D array of blocks

type Inode struct {
	InodeNumber   int         // Inode number
	FileName      [12]byte    // File name
	IsValid       bool        // Set to true if the inode is valid
	IsDirectory   bool        // Set to true for a directory
	ReadWriteLoc  int         // Read/Write location in the file
	DataBlocks    []int       // three direct blocks and one indirect block
	IndirectBlock int         // Indirect block number
	DataSize      int         // Size of the data in the inode
	CreatedTime   time.Time   // Time the inode was created
	LastModified  time.Time   // Time the inode was last modified
	Entries       []FileEntry // Entries in the directory
}

type FileEntry struct {
	FileName [12]byte
	Inode    int
}

type SuperBlock struct {
	InodeStart           int
	FreeBlockBitmapStart int
	DataBlockStart       int
}

var Inodes [MAX_INODES]Inode // Array of inodes

func InitializeFileSystem() {
	initializeSuperBlock()
	sblock := ReadSuperBlock()
	fmt.Println(sblock)
	for i := range Inodes {
		Inodes[i] = Inode{
			IsValid: false,
		}
	}
	var filename [12]byte
	copy(filename[:], "root")
	rootDirectory := Inode{
		InodeNumber:   2,
		FileName:      filename,
		IsValid:       true,
		IsDirectory:   true, // Set to true for a directory
		DataBlocks:    []int{},
		IndirectBlock: 0, // Indirect block number
		CreatedTime:   time.Now(),
		LastModified:  time.Now(),
	}

	// Write the root directory to the disk using inode number 2
	WriteInode(2, rootDirectory)
	fmt.Println("Root directory created as inode 2")
}

func initializeSuperBlock() {
	superBlock := SuperBlock{
		InodeStart:           1,
		FreeBlockBitmapStart: 2,
		DataBlockStart:       7,
	}
	superBlockBytes := EncodeToBytes(superBlock)
	copy(VirtualDisk[0][:], superBlockBytes)
}

func ReadSuperBlock() SuperBlock {
	sBlock := SuperBlock{}
	decoder := gob.NewDecoder(bytes.NewReader(VirtualDisk[0][:]))
	err := decoder.Decode(&sBlock)
	if err != nil {
		log.Fatal("Unable to Decode superblock - better blue Screen", err)
	}
	return sBlock
}

// from https://gist.github.com/SteveBate/042960baa7a4795c3565
func EncodeToBytes(p interface{}) []byte {

	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(p)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("uncompressed size (bytes): ", len(buf.Bytes()))
	return buf.Bytes()
}

func Open(filename string, directory Inode, mode int) Inode {
	if mode == CREATE {
		fmt.Println("File created")
		return CreateFile(filename, directory)
	} else if mode == WRITE {
		fmt.Println("File written")
		return GetCorrectInode(filename, directory)
	} else if mode == APPEND {
		file := GetCorrectInode(filename, directory)
		file.ReadWriteLoc = file.DataSize
		file.LastModified = time.Now()
		return file
	}
	GetCorrectInode(filename, directory)
	ReadFile(Inode{}.InodeNumber)
	// fmt.Println("File Read: ", filename)
	// fmt.Println("Directory: ", directory)
	// fmt.Println("File Data: ", filedata)
	return Inode{}

}

func findInode(fileName string, parentDirectoryInode int) int {
	// Iterate through the directory entries in the parent directory
	for i := 0; i < len(VirtualDisk); i++ {
		directory := Read(i)
		if directory.InodeNumber == parentDirectoryInode && bytes.Equal(directory.FileName[:], []byte(fileName)) {
			return directory.InodeNumber
		}
	}
	return -1
}

func GetCorrectInode(filename string, directory Inode) Inode {
	for _, entry := range Read(directory.InodeNumber).Entries {
		if string(entry.FileName[:]) == filename {
			return ReadInode(entry.Inode)
		}
	}
	return Inode{}
}

func Read(inodeNumber int) Inode {
	// Read the byte array from the disk
	buf := bytes.NewBuffer(VirtualDisk[inodeNumber][:])

	// Convert the byte array to an inode
	inode := Inode{}
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&inode)
	if err != nil {
		log.Fatal("gob.Decode failed:", err)
	}
	// Print the inode
	fmt.Println(inode)

	// Return the inode
	return inode
}

func Write(inodeNumber int, inode Inode) {
	// Convert the inode to a byte array
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(inode)
	if err != nil {
		log.Fatal("gob.Encode failed:", err)
	}

	// Write the byte array to the disk
	copy(VirtualDisk[inodeNumber][:], buf.Bytes())
}

func CreateFile(filename string, parentDirectory Inode) Inode {
	// Find a free inode
	inodeNumber := findFreeInode()

	// Create an inode for the new file
	inode := Inode{
		FileName:      [12]byte{},
		IsValid:       true,
		IsDirectory:   false,
		DataBlocks:    []int{},
		IndirectBlock: 0,
		CreatedTime:   time.Now(),
		LastModified:  time.Now(),
	}

	// Copy the filename to the inode
	copy(inode.FileName[:], filename)

	// Write the inode to the disk
	Write(inodeNumber, inode)

	// Add the new file to the parent directory
	//AddFileToDirectory(filename, inodeNumber, parentDirectory.InodeNumber)

	fmt.Println("File created as inode", inodeNumber)

	// Return the inode number of the new file
	return inode
}

func findFreeInode() int {
	for i, inode := range Inodes {
		if !inode.IsValid {
			return i
		}
	}
	return -1
}

func AddFileToDirectory(filename string, inodeNumber int, parentDirectoryInode int) {
	// Read the parent directory's inode from the disk
	parentDirectory := ReadInode(parentDirectoryInode)

	// Check if the parent directory is actually a directory
	if !parentDirectory.IsDirectory {
		log.Fatal("Inode is not a directory")
	}

	// Create a new file entry
	newFileEntry := FileEntry{
		FileName: [12]byte{},
		Inode:    inodeNumber,
	}
	copy(newFileEntry.FileName[:], filename)

	// Append the new file entry to the parent directory
	parentDirectory.Entries = append(parentDirectory.Entries, newFileEntry)

	// Write the parent directory back to the disk
	WriteInode(parentDirectoryInode, parentDirectory)
}

func ReadInode(inode int) Inode {
	decoder := gob.NewDecoder(bytes.NewReader(VirtualDisk[inode][:]))
	inodeData := Inode{}
	err := decoder.Decode(&inodeData)
	if err != nil {
		log.Fatal("Unable to decode inode", err)
	}
	return inodeData
}

func ReadDataBlock(block int) []byte {
	return VirtualDisk[block][:]
}

func WriteDataBlock(block int, data []byte) {
	copy(VirtualDisk[block][:], data)
}

func WriteInode(inode int, inodeData Inode) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(inodeData)
	if err != nil {
		log.Fatal("Unable to encode inode", err)
	}
	copy(VirtualDisk[inode][:], buf.Bytes())
}

func Unlink(fileNameOrInode string, parentDirectoryInode int) {
	var inodeNumber int
	if fileNameOrInode == "root" {
		inodeNumber = 2
	} else {
		inodeNumber = findInode(fileNameOrInode, parentDirectoryInode)
	}
	inode := ReadInode(inodeNumber)
	inode.IsValid = false
	WriteInode(inodeNumber, inode)
}

func WriteFile(filename string, inodeNumber int, data []byte) {
	// Get the inode from the inode array
	inode := Inodes[inodeNumber]

	// Write the data to the data blocks
	for start := 0; start < len(data); start += BLOCK_SIZE {
		end := start + BLOCK_SIZE
		if end > len(data) {
			end = len(data)
		}

		// Find a free block
		freeBlock := findFreeDataBlock()

		// Write the data to the free block
		copy(VirtualDisk[freeBlock][:], data[start:end])

		// Update the inode's data blocks array
		if len(inode.DataBlocks) < MAX_DIRECT_BLOCKS {
			inode.DataBlocks = append(inode.DataBlocks, freeBlock)
		} else {
			if inode.IndirectBlock == 0 {
				inode.IndirectBlock = findFreeDataBlock()

				// Initialize the indirect block with a slice of zeros
				indirectBlock := make([]int, MAX_INDIRECT_BLOCKS)
				for i := range indirectBlock {
					indirectBlock[i] = 0
				}

				// Convert the slice to a byte array and store it in the virtual disk
				buf := new(bytes.Buffer)
				enc := gob.NewEncoder(buf)
				err := enc.Encode(indirectBlock)
				if err != nil {
					log.Fatal("gob.Encode failed:", err)
				}
				copy(VirtualDisk[inode.IndirectBlock][:], buf.Bytes())
			}

			// Update the indirect block with the free block
			indirectBlock := make([]int, MAX_INDIRECT_BLOCKS)
			buf := bytes.NewBuffer(VirtualDisk[inode.IndirectBlock][:])
			dec := gob.NewDecoder(buf)
			err := dec.Decode(&indirectBlock)
			if err != nil {
				log.Fatal("gob.Decode failed:", err)
			}
			for i, block := range indirectBlock {
				if block == 0 {
					indirectBlock[i] = freeBlock
					break
				}
			}
			buf.Reset()
			enc := gob.NewEncoder(buf)
			err = enc.Encode(indirectBlock)
			if err != nil {
				log.Fatal("gob.Encode failed:", err)
			}
			copy(VirtualDisk[inode.IndirectBlock][:], buf.Bytes())
		}
	}

	// Write the inode back to the inode array
	Inodes[inodeNumber] = inode
}

func ReadFile(inodeNumber int) []byte {
	// Get the inode from the inode array
	inode := Inodes[inodeNumber]

	// Create a byte slice to hold the data
	var data []byte

	// Read the data from each direct data block
	for _, blockNumber := range inode.DataBlocks {
		// Read the block
		blockData := VirtualDisk[blockNumber][:]

		// Append the data from this block to the data slice
		data = append(data, blockData...)
	}

	// If there is an indirect block, read the data from each indirect data block
	if inode.IndirectBlock != 0 {
		// Decode the indirect block
		indirectBlock := make([]int, MAX_INDIRECT_BLOCKS)
		buf := bytes.NewBuffer(VirtualDisk[inode.IndirectBlock][:])
		dec := gob.NewDecoder(buf)
		err := dec.Decode(&indirectBlock)
		if err != nil {
			log.Fatal("gob.Decode failed:", err)
		}

		// Read the data from each indirect data block
		for _, blockNumber := range indirectBlock {
			if blockNumber != 0 {
				// Read the block
				blockData := VirtualDisk[blockNumber][:]

				// Append the data from this block to the data slice
				data = append(data, blockData...)
			}
		}
	}

	fmt.Println("Data read from file:", string(data))
	return data
}

func findFreeDataBlock() int {
	for i, block := range VirtualDisk {
		// Check if the block is free
		if isBlockFree(block) {
			return i
		}
	}

	// Return -1 if no free block is found
	return -1
}

func isBlockFree(block [BLOCK_SIZE]byte) bool {
	for _, b := range block {
		if b != 0 {
			return false
		}
	}
	return true
}
