package key

import (
	"strconv"
	"strings"
)

const (
	maxKeyLength     = 32
	nullIndexPointer = -1
)

//
/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-
//

// Statistic contains the results of the Statistics scan
//
type Statistic struct {
	Active  int
	Deleted int
	Depth   int
	NodeR   int
	NodeS   int
	NodeX   int
	NodeK   int
	NodeL   int
	NodeD   int
}

type indexNode struct {
	status       byte
	leftPointer  int
	key          byte
	rightPointer int
}

//
/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-
//

// Index contains a complete index structure -- one of these is required for each separate index to an array
//
type Index struct {
	indexRoot   int
	node        []indexNode
	deletedRoot int
}

//
/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-
//

// Initialise sets up an index structure
//
func Initialise(indexStructure *Index) {
	indexStructure.indexRoot = nullIndexPointer
	indexStructure.deletedRoot = nullIndexPointer
	return
}

//
/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-
//

func pushStack(inStack []int, keyPointer, inDepth int) (outStack []int, outDepth int) {
	if len(inStack) <= inDepth {
		outStack = append(inStack, keyPointer)
	} else {
		outStack = inStack
		outStack[inDepth] = keyPointer
	}
	outDepth = inDepth + 1
	return
}

//
/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-
//

func decimaliseNumber(keyNumber int) (keyField string, keyLength int) {
	keyField = strconv.Itoa(keyNumber)
	keyLength = len(keyField)
	return
}

//
/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-
//

func extend(keyField string, keyNumber, nextBranchBasePointer int, indexStructure *Index) (extensionPointer int) {
	// creates key nodes in reverse order -- points the leaf node at the next branch sent in as a parameter
	var new indexNode
	var newIndexNumber int
	//
	keyLength := len(keyField)
	extensionPointer = nextBranchBasePointer
	//
	for i := keyLength - 1; i >= 0; i-- {
		// create new key node
		byteArray := []byte(keyField[i : i+1])
		new.key = byteArray[0]
		new.rightPointer = extensionPointer
		if i == keyLength-1 { // leaf node
			new.status = 'S'
			new.leftPointer = keyNumber
		} else { // character node
			new.status = 'X'
			new.leftPointer = nullIndexPointer
		}
		//
		if indexStructure.deletedRoot == nullIndexPointer {
			// append to end of array
			indexStructure.node = append(indexStructure.node, new)
			newIndexNumber = len(indexStructure.node) - 1
		} else {
			// use up one of the "deleted" nodes
			newIndexNumber = indexStructure.deletedRoot
			indexStructure.deletedRoot = indexStructure.node[newIndexNumber].rightPointer
			indexStructure.node[newIndexNumber] = new
		}
		extensionPointer = newIndexNumber
	}
	//
	return
}

//
/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-
//

// Search returns an array of zero or many index numbers based on the input search string and  "searchPrecisely"
// searchPrecisely TRUE -- precise search, key must match existing -- may return many items if duplicates exist
// 			  	   FALSE -- global search, the input string must be a root of a set of one or more existing keys
// (a blank, but not null, input string and "false" will return the entire index in ascending order)
// "matchFound" is "true" if something is located
//
func Search(keyInput string, searchPrecisely bool, indexStructure *Index) (matchFound bool, indexes []int) {
	var keyField string
	var lastMatchPointer int
	//
	if indexStructure.indexRoot == nullIndexPointer { // no index available
		return
	}
	//
	keyInput = strings.TrimSpace(keyInput)
	if len(keyInput) >= maxKeyLength {
		keyField = keyInput[0:maxKeyLength]
	} else {
		keyField = keyInput
	}
	keyLength := len(keyField)
	if keyLength == 0 && searchPrecisely { // nothing to look for
		return
	}
	//
	keyPointer := indexStructure.indexRoot // assume a very
	lastMatchPointer = nullIndexPointer    // wide range
	searching := true
	matchFound = false
	i := 0
	if keyLength == 0 { //nothing to look for so skip searching loop and assume all the tree
		searching = false
		matchFound = true
	}
	//
	for searching { // start searching
		if keyPointer == nullIndexPointer { // looked through it all and didn't find it
			matchFound = false
			searching = false
			break
		}
		switch indexStructure.node[keyPointer].status {
		//
		case 'R', 'S':
			if keyField[i:i+1] == string(indexStructure.node[keyPointer].key) {
				if i+1 == keyLength { // last character in key
					if searchPrecisely { //  narrow the range to just this node
						lastMatchPointer = indexStructure.node[keyPointer].rightPointer
					}
					matchFound = true
					searching = false
					break
				} else { // more characters remain in key
					if indexStructure.node[keyPointer].status == 'S' { // at the terminal leaf so it doesn't match
						matchFound = false
						searching = false
						break
					} // keep traversing
					keyPointer = indexStructure.node[keyPointer].rightPointer
					i++
				}
			} else { // key doesn't match
				matchFound = false
				searching = false
				break
			}
		//
		case 'D':
			if keyField[i:i+1] <= string(indexStructure.node[keyPointer].key) {
				if !searchPrecisely { // global search
					lastMatchPointer = keyPointer //  keep track of the base of the "current" branch
				}
				keyPointer = indexStructure.node[keyPointer].leftPointer
			} else {
				keyPointer = indexStructure.node[keyPointer].rightPointer
			}
		//
		case 'K', 'L':
			if keyField[i:i+1] == string(indexStructure.node[keyPointer].key) {
				if i+1 == keyLength { // last character in key
					if searchPrecisely { //  narrow the range
						lastMatchPointer = keyPointer
					}
					keyPointer = indexStructure.node[keyPointer].leftPointer //  move into the duplicate branch
					matchFound = true
					searching = false
					break
				} else { // more characters remain in key
					if indexStructure.node[keyPointer].status == 'L' { // at the terminal leaf so it doesn't match
						matchFound = false
						searching = false
						break
					} // keep traversing
					keyPointer = indexStructure.node[keyPointer].rightPointer
					i++
				}
			} else { // key doesn't match
				matchFound = false
				searching = false
				break
			}
		//
		case 'X':
			if keyField[i:i+1] == string(indexStructure.node[keyPointer].key) {
				if i+1 == keyLength { // last character in key
					if searchPrecisely {
						matchFound = false
					} else { // global search
						keyPointer = indexStructure.node[keyPointer].rightPointer
						matchFound = true
					}
					searching = false
					break
				} else { // more characters remain in key so keep traversing
					keyPointer = indexStructure.node[keyPointer].rightPointer
					i++
				}
			} else { // key doesn't match
				matchFound = false
				searching = false
				break
			}
		}
	} // end searching
	//
	if matchFound {
		// scan the tree and collect results -- from keyPointer to (lastMatchPointer or -1)
		goLeftAtNextNode := true
		for keyPointer != lastMatchPointer && keyPointer != nullIndexPointer {
			switch indexStructure.node[keyPointer].status {
			//
			case 'R':
				indexes = append(indexes, indexStructure.node[keyPointer].leftPointer)
				keyPointer = indexStructure.node[keyPointer].rightPointer
				goLeftAtNextNode = true
			//
			case 'S':
				indexes = append(indexes, indexStructure.node[keyPointer].leftPointer)
				keyPointer = indexStructure.node[keyPointer].rightPointer
				goLeftAtNextNode = false
			//
			case 'K':
				if goLeftAtNextNode {
					keyPointer = indexStructure.node[keyPointer].leftPointer
				} else {
					keyPointer = indexStructure.node[keyPointer].rightPointer
				}
				goLeftAtNextNode = true
			//
			case 'L':
				if goLeftAtNextNode {
					keyPointer = indexStructure.node[keyPointer].leftPointer
				} else {
					keyPointer = indexStructure.node[keyPointer].rightPointer
				}
			//
			case 'D':
				if goLeftAtNextNode {
					keyPointer = indexStructure.node[keyPointer].leftPointer
				} else {
					keyPointer = indexStructure.node[keyPointer].rightPointer
				}
				goLeftAtNextNode = true
			//
			case 'X':
				keyPointer = indexStructure.node[keyPointer].rightPointer
				goLeftAtNextNode = true
			}
		}
	}
	return
}

//
/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-
//

// Delete removes a key and its associated "index-number" from the supplied index
//
func Delete(keyInput string, keyNumber int, indexStructure *Index) {

	var keyField string
	//
	keyInput = strings.TrimSpace(keyInput)
	if len(keyInput) >= maxKeyLength {
		keyField = keyInput[0:maxKeyLength]
	} else {
		keyField = keyInput
	}
	keyLength := len(keyField)
	if keyLength == 0 {
		return
	}
	if indexStructure.indexRoot == nullIndexPointer { // no index
		return
	}
	keyPointer := indexStructure.indexRoot
	previousIndexNumber := nullIndexPointer
	linkIndexNumber := nullIndexPointer
	deleteIndexNumber := nullIndexPointer
	duplicateIndexNumber := nullIndexPointer
	//
	goLeft := true
	duplicateCount := 0
	i := 0
	for searching := true; searching; { // start searching
		//
		if indexStructure.node[keyPointer].status == 'D' ||
			indexStructure.node[keyPointer].status == 'K' ||
			indexStructure.node[keyPointer].status == 'R' {
			deleteIndexNumber = keyPointer
			linkIndexNumber = previousIndexNumber
		}
		//
		switch indexStructure.node[keyPointer].status {
		//
		case 'R', 'S':
			if keyField[i:i+1] == string(indexStructure.node[keyPointer].key) {
				if duplicateIndexNumber != nullIndexPointer && indexStructure.node[keyPointer].status == 'R' {
					duplicateCount++
				}
				if i+1 == keyLength {
					if indexStructure.node[keyPointer].leftPointer != keyNumber {
						return
					}
					searching = false
					break
				} else {
					if indexStructure.node[keyPointer].status == 'S' {
						return
					}
					previousIndexNumber = keyPointer
					keyPointer = indexStructure.node[keyPointer].rightPointer
					i++
				}
			} else {
				return
			}
		//
		case 'D':
			if duplicateIndexNumber != nullIndexPointer {
				duplicateCount++
			}
			previousIndexNumber = keyPointer
			if keyField[i:i+1] <= string(indexStructure.node[keyPointer].key) {
				keyPointer = indexStructure.node[keyPointer].leftPointer
				goLeft = true
			} else {
				keyPointer = indexStructure.node[keyPointer].rightPointer
				goLeft = false
			}
			//
		case 'K', 'L':
			if keyField[i:i+1] == string(indexStructure.node[keyPointer].key) {
				previousIndexNumber = keyPointer
				if i+1 == keyLength {
					duplicateIndexNumber = keyPointer
					i = 0
					keyField, keyLength = decimaliseNumber(keyNumber)
					keyPointer = indexStructure.node[keyPointer].leftPointer
				} else {
					if indexStructure.node[keyPointer].status == 'L' {
						return
					}
					keyPointer = indexStructure.node[keyPointer].rightPointer
					i++
				}
			} else {
				return
			}
			//
		case 'X':
			if keyField[i:i+1] == string(indexStructure.node[keyPointer].key) {
				if i+1 == keyLength {
					return
				}
				previousIndexNumber = keyPointer
				keyPointer = indexStructure.node[keyPointer].rightPointer
				i++
			} else {
				return
			}
		}
	} // end searching
	//
	if duplicateIndexNumber != nullIndexPointer { // duplicate tree found
		//
		if duplicateCount == 0 { // should never happen
			indexStructure.node[keyPointer].rightPointer = indexStructure.deletedRoot
			indexStructure.deletedRoot = indexStructure.node[duplicateIndexNumber].leftPointer
			if indexStructure.node[duplicateIndexNumber].status == 'K' {
				indexStructure.node[duplicateIndexNumber].status = 'X'
				indexStructure.node[duplicateIndexNumber].leftPointer = nullIndexPointer
			} else {
				indexStructure.node[duplicateIndexNumber].status = 'S'
				indexStructure.node[duplicateIndexNumber].leftPointer = keyNumber
			}
			if indexStructure.node[duplicateIndexNumber].status == 'X' {
				return
			}
			keyPointer = duplicateIndexNumber
			duplicateIndexNumber = nullIndexPointer
			//
		} else {
			var tempIndexNumber int
			if duplicateCount == 1 { // scan up the tree looking for branches ('D') and keys ('R')
				if indexStructure.node[keyPointer].status == 'R' {
					tempIndexNumber = indexStructure.node[keyPointer].rightPointer
				} else if indexStructure.node[deleteIndexNumber].status == 'R' || goLeft {
					tempIndexNumber = indexStructure.node[deleteIndexNumber].rightPointer
				} else {
					tempIndexNumber = indexStructure.node[deleteIndexNumber].leftPointer
				}
				for searching := true; searching; {
					if indexStructure.node[tempIndexNumber].status == 'R' ||
						(indexStructure.node[tempIndexNumber].status == 'D' &&
							tempIndexNumber != deleteIndexNumber) {
						duplicateCount++
					}
					if indexStructure.node[tempIndexNumber].rightPointer == duplicateIndexNumber ||
						duplicateCount > 1 {
						searching = false
					} else {
						tempIndexNumber = indexStructure.node[tempIndexNumber].rightPointer
					}
				}
			}
			if duplicateCount == 1 { // still only got one duplicate after scanning up the tree
				if indexStructure.node[keyPointer].status == 'R' {
					indexStructure.node[tempIndexNumber].rightPointer = indexStructure.deletedRoot
					indexStructure.deletedRoot = indexStructure.node[duplicateIndexNumber].leftPointer
					if indexStructure.node[duplicateIndexNumber].status == 'K' {
						indexStructure.node[duplicateIndexNumber].status = 'R'
					} else {
						indexStructure.node[duplicateIndexNumber].status = 'S'
					}
					indexStructure.node[duplicateIndexNumber].leftPointer =
						indexStructure.node[tempIndexNumber].leftPointer
					return
				}
				if indexStructure.node[linkIndexNumber].status == 'K' {
					indexStructure.node[linkIndexNumber].status = 'R'
				} else {
					indexStructure.node[linkIndexNumber].status = 'S'
				}
				if indexStructure.node[deleteIndexNumber].status == 'R' {
					indexStructure.node[keyPointer].rightPointer = indexStructure.deletedRoot
					indexStructure.deletedRoot = indexStructure.node[linkIndexNumber].leftPointer
					indexStructure.node[linkIndexNumber].leftPointer =
						indexStructure.node[deleteIndexNumber].leftPointer

				} else {
					saveIndex := indexStructure.node[linkIndexNumber].leftPointer
					var resetRightIndex, endIndex int
					if goLeft {
						resetRightIndex = indexStructure.node[deleteIndexNumber].rightPointer
						endIndex = duplicateIndexNumber
					} else {
						resetRightIndex = indexStructure.node[deleteIndexNumber].leftPointer
						endIndex = deleteIndexNumber
					}
					for indexStructure.node[resetRightIndex].rightPointer != endIndex {
						resetRightIndex = indexStructure.node[resetRightIndex].rightPointer
					}
					indexStructure.node[linkIndexNumber].leftPointer =
						indexStructure.node[resetRightIndex].leftPointer
					if goLeft {
						indexStructure.node[keyPointer].rightPointer = indexStructure.deletedRoot
						indexStructure.node[resetRightIndex].rightPointer =
							indexStructure.node[deleteIndexNumber].leftPointer
					} else {
						indexStructure.node[resetRightIndex].rightPointer = indexStructure.deletedRoot
						indexStructure.node[keyPointer].rightPointer =
							indexStructure.node[deleteIndexNumber].leftPointer
					}
					indexStructure.deletedRoot = saveIndex
				}
				return
			}
		}
	} // end duplicate tree
	//
	if indexStructure.node[keyPointer].status == 'R' {
		indexStructure.node[keyPointer].status = 'X'
		indexStructure.node[keyPointer].leftPointer = nullIndexPointer
		return
	}
	//
	if deleteIndexNumber == nullIndexPointer {
		indexStructure.node[keyPointer].rightPointer = indexStructure.deletedRoot
		indexStructure.deletedRoot = indexStructure.indexRoot
		indexStructure.indexRoot = nullIndexPointer
		return
	}
	//
	if indexStructure.node[deleteIndexNumber].status == 'R' || indexStructure.node[deleteIndexNumber].status == 'K' {
		saveIndex := indexStructure.node[deleteIndexNumber].rightPointer
		indexStructure.node[deleteIndexNumber].rightPointer = indexStructure.node[keyPointer].rightPointer
		if indexStructure.node[deleteIndexNumber].status == 'R' {
			indexStructure.node[deleteIndexNumber].status = 'S'
		} else {
			indexStructure.node[deleteIndexNumber].status = 'L'
		}
		indexStructure.node[keyPointer].rightPointer = indexStructure.deletedRoot
		indexStructure.deletedRoot = saveIndex
		return
	}
	//
	if goLeft {
		if linkIndexNumber == nullIndexPointer {
			indexStructure.indexRoot = indexStructure.node[deleteIndexNumber].rightPointer
		} else {
			if indexStructure.node[linkIndexNumber].status == 'D' {
				if indexStructure.node[deleteIndexNumber].key <= indexStructure.node[linkIndexNumber].key {
					resetIndex := indexStructure.node[deleteIndexNumber].rightPointer
					if indexStructure.node[resetIndex].status != 'D' {
						indexStructure.node[linkIndexNumber].key = indexStructure.node[resetIndex].key
					}
					indexStructure.node[linkIndexNumber].leftPointer =
						indexStructure.node[deleteIndexNumber].rightPointer
				} else {
					indexStructure.node[linkIndexNumber].rightPointer =
						indexStructure.node[deleteIndexNumber].rightPointer
				}
			} else if (indexStructure.node[linkIndexNumber].status == 'K' ||
				indexStructure.node[linkIndexNumber].status == 'L') && duplicateIndexNumber != nullIndexPointer {
				indexStructure.node[linkIndexNumber].leftPointer = indexStructure.node[deleteIndexNumber].rightPointer
			} else {
				indexStructure.node[linkIndexNumber].rightPointer = indexStructure.node[deleteIndexNumber].rightPointer
			}
		}
		indexStructure.node[keyPointer].rightPointer = indexStructure.deletedRoot
		indexStructure.deletedRoot = deleteIndexNumber
		indexStructure.node[deleteIndexNumber].rightPointer = indexStructure.node[deleteIndexNumber].leftPointer
	} else {
		threadIndex := indexStructure.node[deleteIndexNumber].leftPointer
		for indexStructure.node[threadIndex].rightPointer != deleteIndexNumber {
			threadIndex = indexStructure.node[threadIndex].rightPointer
		}
		indexStructure.node[threadIndex].rightPointer = indexStructure.node[keyPointer].rightPointer
		if linkIndexNumber == nullIndexPointer {
			indexStructure.indexRoot = indexStructure.node[deleteIndexNumber].leftPointer
		} else {
			if indexStructure.node[linkIndexNumber].status == 'D' {
				if indexStructure.node[deleteIndexNumber].key <= indexStructure.node[linkIndexNumber].key {
					resetIndex := indexStructure.node[deleteIndexNumber].leftPointer
					if indexStructure.node[resetIndex].status != 'D' {
						indexStructure.node[linkIndexNumber].key = indexStructure.node[resetIndex].key
					}
					indexStructure.node[linkIndexNumber].leftPointer =
						indexStructure.node[deleteIndexNumber].leftPointer
				} else {
					indexStructure.node[linkIndexNumber].rightPointer =
						indexStructure.node[deleteIndexNumber].leftPointer
				}
			} else if (indexStructure.node[linkIndexNumber].status == 'K' ||
				indexStructure.node[linkIndexNumber].status == 'L') && duplicateIndexNumber != nullIndexPointer {
				indexStructure.node[linkIndexNumber].leftPointer = indexStructure.node[deleteIndexNumber].leftPointer
			} else {
				indexStructure.node[linkIndexNumber].rightPointer = indexStructure.node[deleteIndexNumber].leftPointer
			}
		}
		indexStructure.node[keyPointer].rightPointer = indexStructure.deletedRoot
		indexStructure.deletedRoot = deleteIndexNumber
	}
	//
}

//
/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-
//

// Insert places the input string into the specified index structure along with the supplied "index-number"
//
func Insert(keyInput string, keyNumber int, indexStructure *Index) {
	var keyField string
	var decisionIndexNumber, lastIndexNumber int
	//
	keyInput = strings.TrimSpace(keyInput)
	if len(keyInput) >= maxKeyLength {
		keyField = keyInput[0:maxKeyLength]
	} else {
		keyField = keyInput
	}
	keyLength := len(keyField)
	if keyLength == 0 {
		return
	}
	if indexStructure.indexRoot == nullIndexPointer { // no index so just put the key straight into the structure
		indexStructure.indexRoot = extend(keyField, keyNumber, nullIndexPointer, indexStructure)
		return
	}
	keyPointer := indexStructure.indexRoot
	previousIndexNumber := nullIndexPointer
	duplicateFlag := false
	i := 0
	//
	for searching := true; searching; { // start searching
		switch indexStructure.node[keyPointer].status {
		//
		case 'R', 'S':
			if keyField[i:i+1] == string(indexStructure.node[keyPointer].key) {
				if i+1 == keyLength {
					if keyNumber == indexStructure.node[keyPointer].leftPointer { // new key is EXACTLY same as existing
						return // key value and key number are the same so do nothing
					}
					keyField, keyLength = decimaliseNumber(indexStructure.node[keyPointer].leftPointer)
					linkIndexNumber :=
						extend(keyField, indexStructure.node[keyPointer].leftPointer, keyPointer, indexStructure)
					if indexStructure.node[keyPointer].status == 'R' {
						indexStructure.node[keyPointer].status = 'K'
					} else {
						indexStructure.node[keyPointer].status = 'L'
					}
					indexStructure.node[keyPointer].leftPointer = linkIndexNumber
					duplicateFlag = true
					i = 0
					keyField, keyLength = decimaliseNumber(keyNumber)
					previousIndexNumber = keyPointer
					keyPointer = linkIndexNumber
				} else {
					if indexStructure.node[keyPointer].status == 'S' {
						searching = false
						break
					}
					previousIndexNumber = keyPointer
					keyPointer = indexStructure.node[keyPointer].rightPointer
					i++
				}
			} else {
				searching = false
				break
			}
		//
		case 'D':
			previousIndexNumber = keyPointer
			if keyField[i:i+1] <= string(indexStructure.node[keyPointer].key) {
				keyPointer = indexStructure.node[keyPointer].leftPointer
			} else {
				keyPointer = indexStructure.node[keyPointer].rightPointer
			}
		//
		case 'K', 'L':
			if keyField[i:i+1] == string(indexStructure.node[keyPointer].key) {
				if i+1 == keyLength {
					duplicateFlag = true
					keyField, keyLength = decimaliseNumber(keyNumber) // start a new key
					i = 0
					previousIndexNumber = keyPointer
					keyPointer = indexStructure.node[keyPointer].leftPointer
				} else {
					if indexStructure.node[keyPointer].status == 'L' {
						searching = false
						break
					}
					previousIndexNumber = keyPointer
					keyPointer = indexStructure.node[keyPointer].rightPointer
					i++
				}
			} else {
				searching = false
				break
			}
		//
		case 'X':
			if keyField[i:i+1] == string(indexStructure.node[keyPointer].key) {
				if i+1 == keyLength {
					searching = false
					break
				}
				previousIndexNumber = keyPointer
				keyPointer = indexStructure.node[keyPointer].rightPointer
				i++
			} else {
				searching = false
				break
			}
		}
	} // end searching
	//
	if keyField[i:i+1] == string(indexStructure.node[keyPointer].key) {
		if indexStructure.node[keyPointer].status == 'X' {
			indexStructure.node[keyPointer].status = 'R'
			indexStructure.node[keyPointer].leftPointer = keyNumber
			return
		}
		i++
		linkIndexNumber :=
			extend(keyField[i:], keyNumber, indexStructure.node[keyPointer].rightPointer, indexStructure)
		if indexStructure.node[keyPointer].status == 'S' {
			indexStructure.node[keyPointer].status = 'R'
		} else {
			indexStructure.node[keyPointer].status = 'K' // was an "L" before
		}
		indexStructure.node[keyPointer].rightPointer = linkIndexNumber
		return
	}
	//
	placeholderNode := indexNode{ // make a dummy "D" node as placeholder
		status:       'D',
		leftPointer:  nullIndexPointer,
		key:          ' ',
		rightPointer: nullIndexPointer,
	}
	if indexStructure.deletedRoot == nullIndexPointer {
		// append new decision to end of array
		indexStructure.node = append(indexStructure.node, placeholderNode)
		decisionIndexNumber = len(indexStructure.node) - 1
	} else {
		// use up one of the "deleted" nodes
		decisionIndexNumber = indexStructure.deletedRoot
		indexStructure.deletedRoot = indexStructure.node[decisionIndexNumber].rightPointer
		indexStructure.node[decisionIndexNumber] = placeholderNode
	}
	//
	if keyField[i:i+1] > string(indexStructure.node[keyPointer].key) {
		threadIndex := keyPointer
		for !(indexStructure.node[threadIndex].status == 'L' || indexStructure.node[threadIndex].status == 'S') {
			threadIndex = indexStructure.node[threadIndex].rightPointer
		}
		lastIndexNumber = indexStructure.node[threadIndex].rightPointer
		indexStructure.node[threadIndex].rightPointer = decisionIndexNumber
	} else {
		lastIndexNumber = decisionIndexNumber
	}
	//
	linkIndexNumber := extend(keyField[i:], keyNumber, lastIndexNumber, indexStructure)
	//
	if keyField[i:i+1] < string(indexStructure.node[keyPointer].key) {
		indexStructure.node[decisionIndexNumber].leftPointer = linkIndexNumber
		byteArray := []byte(keyField[i : i+1])
		indexStructure.node[decisionIndexNumber].key = byteArray[0]
		indexStructure.node[decisionIndexNumber].rightPointer = keyPointer
	} else {
		indexStructure.node[decisionIndexNumber].leftPointer = keyPointer
		indexStructure.node[decisionIndexNumber].key = indexStructure.node[keyPointer].key
		indexStructure.node[decisionIndexNumber].rightPointer = linkIndexNumber
	}
	//
	if previousIndexNumber == nullIndexPointer {
		indexStructure.indexRoot = decisionIndexNumber
	} else {
		if indexStructure.node[previousIndexNumber].status == 'D' &&
			keyField[i:i+1] <= string(indexStructure.node[previousIndexNumber].key) ||
			indexStructure.node[previousIndexNumber].status == 'L' ||
			(indexStructure.node[previousIndexNumber].status == 'K' && duplicateFlag) {
			indexStructure.node[previousIndexNumber].leftPointer = decisionIndexNumber
		} else {
			indexStructure.node[previousIndexNumber].rightPointer = decisionIndexNumber
		}
	}
	return
}

//
/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-/////////-
//

// Statistics scans the specified index structure and returns a structure of counts of the different node types
//
func Statistics(indexStructure *Index) (result Statistic) {
	var stack []int
	stackPointer := 0
	goLeft := true
	result.Active = 0
	result.Deleted = 0
	result.Depth = 0
	result.NodeR = 0
	result.NodeS = 0
	result.NodeX = 0
	result.NodeK = 0
	result.NodeL = 0
	result.NodeD = 0
	keyPointer := indexStructure.indexRoot
	//
	for scanning := true; scanning; { // start scanning
		if keyPointer == nullIndexPointer { // looked through it all, or nothing there to begin with
			scanning = false
			break
		}
		switch indexStructure.node[keyPointer].status {
		//
		case 'X':
			result.NodeX++
			stack, stackPointer = pushStack(stack, keyPointer, stackPointer)
			keyPointer = indexStructure.node[keyPointer].rightPointer
			goLeft = true
			//
		case 'R':
			result.NodeR++
			stack, stackPointer = pushStack(stack, keyPointer, stackPointer)
			keyPointer = indexStructure.node[keyPointer].rightPointer
			goLeft = true
			//
		case 'D':
			if goLeft {
				result.NodeD++
				stack, stackPointer = pushStack(stack, keyPointer, stackPointer)
				keyPointer = indexStructure.node[keyPointer].leftPointer
			} else {
				keyPointer = indexStructure.node[keyPointer].rightPointer
				goLeft = true
			}
			//
		case 'K':
			if goLeft {
				result.NodeK++
				stack, stackPointer = pushStack(stack, keyPointer, stackPointer)
				keyPointer = indexStructure.node[keyPointer].leftPointer
			} else {
				keyPointer = indexStructure.node[keyPointer].rightPointer
				goLeft = true
			}
			//
		case 'S':
			result.NodeS++
			stack, stackPointer = pushStack(stack, keyPointer, stackPointer)
			keyPointer = indexStructure.node[keyPointer].rightPointer
			if keyPointer != nullIndexPointer { // reset the stack
				for stackPointer = 0; stack[stackPointer] != keyPointer; stackPointer++ {
				}
				stackPointer++
			}
			goLeft = false // goRight
			//
		case 'L':
			if goLeft {
				result.NodeL++
				stack, stackPointer = pushStack(stack, keyPointer, stackPointer)
				keyPointer = indexStructure.node[keyPointer].leftPointer
			} else { // going Right and staying Right
				keyPointer = indexStructure.node[keyPointer].rightPointer
				if keyPointer != nullIndexPointer { // reset the stack
					for stackPointer = 0; stack[stackPointer] != keyPointer; stackPointer++ {
					}
					stackPointer++
				}
			}
		}
	} // end scanning
	//
	result.Active = result.NodeR + result.NodeS + result.NodeX + result.NodeK + result.NodeL + result.NodeD
	result.Depth = len(stack)
	for x := indexStructure.deletedRoot; x != nullIndexPointer; x = indexStructure.node[x].rightPointer {
		result.Deleted++
	}
	return
}
