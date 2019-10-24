package conntrack

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"math"
)

const shiftSize = 5
const nodeSize = 32
const shiftBitMask = 0x1F

type commonNode interface{}

var emptyCommonNode commonNode = []commonNode{}

func uintMin(a, b uint) uint {
	if a < b {
		return a
	}

	return b
}

func newPath(shift uint, node commonNode) commonNode {
	if shift == 0 {
		return node
	}

	return newPath(shift-shiftSize, commonNode([]commonNode{node}))
}

func assertSliceOk(start, stop, len int) {
	if start < 0 {
		panic(fmt.Sprintf("Invalid slice index %d (index must be non-negative)", start))
	}

	if start > stop {
		panic(fmt.Sprintf("Invalid slice index: %d > %d", start, stop))
	}

	if stop > len {
		panic(fmt.Sprintf("Slice bounds out of range, start=%d, stop=%d, len=%d", start, stop, len))
	}
}

const upperMapLoadFactor float64 = 8.0
const lowerMapLoadFactor float64 = 2.0
const initialMapLoadFactor float64 = (upperMapLoadFactor + lowerMapLoadFactor) / 2

//////////////////////////
//// Hash functions //////
//////////////////////////

func hash(x []byte) uint32 {
	return crc32.ChecksumIEEE(x)
}

func interfaceHash(x interface{}) uint32 {
	return hash([]byte(fmt.Sprintf("%v", x)))
}

func byteHash(x byte) uint32 {
	return hash([]byte{x})
}

func uint8Hash(x uint8) uint32 {
	return byteHash(byte(x))
}

func int8Hash(x int8) uint32 {
	return uint8Hash(uint8(x))
}

func uint16Hash(x uint16) uint32 {
	bX := make([]byte, 2)
	binary.LittleEndian.PutUint16(bX, x)
	return hash(bX)
}

func int16Hash(x int16) uint32 {
	return uint16Hash(uint16(x))
}

func uint32Hash(x uint32) uint32 {
	bX := make([]byte, 4)
	binary.LittleEndian.PutUint32(bX, x)
	return hash(bX)
}

func int32Hash(x int32) uint32 {
	return uint32Hash(uint32(x))
}

func uint64Hash(x uint64) uint32 {
	bX := make([]byte, 8)
	binary.LittleEndian.PutUint64(bX, x)
	return hash(bX)
}

func int64Hash(x int64) uint32 {
	return uint64Hash(uint64(x))
}

func intHash(x int) uint32 {
	return int64Hash(int64(x))
}

func uintHash(x uint) uint32 {
	return uint64Hash(uint64(x))
}

func boolHash(x bool) uint32 {
	if x {
		return 1
	}

	return 0
}

func runeHash(x rune) uint32 {
	return int32Hash(int32(x))
}

func stringHash(x string) uint32 {
	return hash([]byte(x))
}

func float64Hash(x float64) uint32 {
	return uint64Hash(math.Float64bits(x))
}

func float32Hash(x float32) uint32 {
	return uint32Hash(math.Float32bits(x))
}

///////////
/// Map ///
///////////

//////////////////////
/// Backing vector ///
//////////////////////

type privateprivateentryHandleSetMapItemBucketVector struct {
	tail  []privateprivateentryHandleSetMapItemBucket
	root  commonNode
	len   uint
	shift uint
}

type privateentryHandleSetMapItem struct {
	Key   *EntryHandle
	Value struct{}
}

type privateprivateentryHandleSetMapItemBucket []privateentryHandleSetMapItem

var emptyprivateentryHandleSetMapItemBucketVectorTail = make([]privateprivateentryHandleSetMapItemBucket, 0)
var emptyprivateentryHandleSetMapItemBucketVector *privateprivateentryHandleSetMapItemBucketVector = &privateprivateentryHandleSetMapItemBucketVector{root: emptyCommonNode, shift: shiftSize, tail: emptyprivateentryHandleSetMapItemBucketVectorTail}

func (v *privateprivateentryHandleSetMapItemBucketVector) Get(i int) privateprivateentryHandleSetMapItemBucket {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	return v.sliceFor(uint(i))[i&shiftBitMask]
}

func (v *privateprivateentryHandleSetMapItemBucketVector) sliceFor(i uint) []privateprivateentryHandleSetMapItemBucket {
	if i >= v.tailOffset() {
		return v.tail
	}

	node := v.root
	for level := v.shift; level > 0; level -= shiftSize {
		node = node.([]commonNode)[(i>>level)&shiftBitMask]
	}

	return node.([]privateprivateentryHandleSetMapItemBucket)
}

func (v *privateprivateentryHandleSetMapItemBucketVector) tailOffset() uint {
	if v.len < nodeSize {
		return 0
	}

	return ((v.len - 1) >> shiftSize) << shiftSize
}

func (v *privateprivateentryHandleSetMapItemBucketVector) Set(i int, item privateprivateentryHandleSetMapItemBucket) *privateprivateentryHandleSetMapItemBucketVector {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	if uint(i) >= v.tailOffset() {
		newTail := make([]privateprivateentryHandleSetMapItemBucket, len(v.tail))
		copy(newTail, v.tail)
		newTail[i&shiftBitMask] = item
		return &privateprivateentryHandleSetMapItemBucketVector{root: v.root, tail: newTail, len: v.len, shift: v.shift}
	}

	return &privateprivateentryHandleSetMapItemBucketVector{root: v.doAssoc(v.shift, v.root, uint(i), item), tail: v.tail, len: v.len, shift: v.shift}
}

func (v *privateprivateentryHandleSetMapItemBucketVector) doAssoc(level uint, node commonNode, i uint, item privateprivateentryHandleSetMapItemBucket) commonNode {
	if level == 0 {
		ret := make([]privateprivateentryHandleSetMapItemBucket, nodeSize)
		copy(ret, node.([]privateprivateentryHandleSetMapItemBucket))
		ret[i&shiftBitMask] = item
		return ret
	}

	ret := make([]commonNode, nodeSize)
	copy(ret, node.([]commonNode))
	subidx := (i >> level) & shiftBitMask
	ret[subidx] = v.doAssoc(level-shiftSize, ret[subidx], i, item)
	return ret
}

func (v *privateprivateentryHandleSetMapItemBucketVector) pushTail(level uint, parent commonNode, tailNode []privateprivateentryHandleSetMapItemBucket) commonNode {
	subIdx := ((v.len - 1) >> level) & shiftBitMask
	parentNode := parent.([]commonNode)
	ret := make([]commonNode, subIdx+1)
	copy(ret, parentNode)
	var nodeToInsert commonNode

	if level == shiftSize {
		nodeToInsert = tailNode
	} else if subIdx < uint(len(parentNode)) {
		nodeToInsert = v.pushTail(level-shiftSize, parentNode[subIdx], tailNode)
	} else {
		nodeToInsert = newPath(level-shiftSize, tailNode)
	}

	ret[subIdx] = nodeToInsert
	return ret
}

func (v *privateprivateentryHandleSetMapItemBucketVector) Append(item ...privateprivateentryHandleSetMapItemBucket) *privateprivateentryHandleSetMapItemBucketVector {
	result := v
	itemLen := uint(len(item))
	for insertOffset := uint(0); insertOffset < itemLen; {
		tailLen := result.len - result.tailOffset()
		tailFree := nodeSize - tailLen
		if tailFree == 0 {
			result = result.pushLeafNode(result.tail)
			result.tail = emptyprivateentryHandleSetMapItemBucketVector.tail
			tailFree = nodeSize
			tailLen = 0
		}

		batchLen := uintMin(itemLen-insertOffset, tailFree)
		newTail := make([]privateprivateentryHandleSetMapItemBucket, 0, tailLen+batchLen)
		newTail = append(newTail, result.tail...)
		newTail = append(newTail, item[insertOffset:insertOffset+batchLen]...)
		result = &privateprivateentryHandleSetMapItemBucketVector{root: result.root, tail: newTail, len: result.len + batchLen, shift: result.shift}
		insertOffset += batchLen
	}

	return result
}

func (v *privateprivateentryHandleSetMapItemBucketVector) pushLeafNode(node []privateprivateentryHandleSetMapItemBucket) *privateprivateentryHandleSetMapItemBucketVector {
	var newRoot commonNode
	newShift := v.shift

	// Root overflow?
	if (v.len >> shiftSize) > (1 << v.shift) {
		newNode := newPath(v.shift, node)
		newRoot = commonNode([]commonNode{v.root, newNode})
		newShift = v.shift + shiftSize
	} else {
		newRoot = v.pushTail(v.shift, v.root, node)
	}

	return &privateprivateentryHandleSetMapItemBucketVector{root: newRoot, tail: v.tail, len: v.len, shift: newShift}
}

func (v *privateprivateentryHandleSetMapItemBucketVector) Len() int {
	return int(v.len)
}

func (v *privateprivateentryHandleSetMapItemBucketVector) Range(f func(privateprivateentryHandleSetMapItemBucket) bool) {
	var currentNode []privateprivateentryHandleSetMapItemBucket
	for i := uint(0); i < v.len; i++ {
		if i&shiftBitMask == 0 {
			currentNode = v.sliceFor(uint(i))
		}

		if !f(currentNode[i&shiftBitMask]) {
			return
		}
	}
}

// privateentryHandleSetMap is a persistent key - value map
type privateentryHandleSetMap struct {
	backingVector *privateprivateentryHandleSetMapItemBucketVector
	len           int
}

func (m *privateentryHandleSetMap) pos(key *EntryHandle) int {
	return int(uint64(interfaceHash(key)) % uint64(m.backingVector.Len()))
}

// Helper type used during map creation and reallocation
type privateprivateentryHandleSetMapItemBuckets struct {
	buckets []privateprivateentryHandleSetMapItemBucket
	length  int
}

func newPrivateprivateentryHandleSetMapItemBuckets(itemCount int) *privateprivateentryHandleSetMapItemBuckets {
	size := int(float64(itemCount)/initialMapLoadFactor) + 1
	buckets := make([]privateprivateentryHandleSetMapItemBucket, size)
	return &privateprivateentryHandleSetMapItemBuckets{buckets: buckets}
}

func (b *privateprivateentryHandleSetMapItemBuckets) AddItem(item privateentryHandleSetMapItem) {
	ix := int(uint64(interfaceHash(item.Key)) % uint64(len(b.buckets)))
	bucket := b.buckets[ix]
	if bucket != nil {
		// Hash collision, merge with existing bucket
		for keyIx, bItem := range bucket {
			if item.Key == bItem.Key {
				bucket[keyIx] = item
				return
			}
		}

		b.buckets[ix] = append(bucket, privateentryHandleSetMapItem{Key: item.Key, Value: item.Value})
		b.length++
	} else {
		bucket := make(privateprivateentryHandleSetMapItemBucket, 0, int(math.Max(initialMapLoadFactor, 1.0)))
		b.buckets[ix] = append(bucket, item)
		b.length++
	}
}

func (b *privateprivateentryHandleSetMapItemBuckets) AddItemsFromMap(m *privateentryHandleSetMap) {
	m.backingVector.Range(func(bucket privateprivateentryHandleSetMapItemBucket) bool {
		for _, item := range bucket {
			b.AddItem(item)
		}
		return true
	})
}

func newprivateentryHandleSetMap(items []privateentryHandleSetMapItem) *privateentryHandleSetMap {
	buckets := newPrivateprivateentryHandleSetMapItemBuckets(len(items))
	for _, item := range items {
		buckets.AddItem(item)
	}

	return &privateentryHandleSetMap{backingVector: emptyprivateentryHandleSetMapItemBucketVector.Append(buckets.buckets...), len: buckets.length}
}

// Len returns the number of items in m.
func (m *privateentryHandleSetMap) Len() int {
	return int(m.len)
}

// Load returns value identified by key. ok is set to true if key exists in the map, false otherwise.
func (m *privateentryHandleSetMap) Load(key *EntryHandle) (value struct{}, ok bool) {
	bucket := m.backingVector.Get(m.pos(key))
	if bucket != nil {
		for _, item := range bucket {
			if item.Key == key {
				return item.Value, true
			}
		}
	}

	var zeroValue struct{}
	return zeroValue, false
}

// Store returns a new privateentryHandleSetMap containing value identified by key.
func (m *privateentryHandleSetMap) Store(key *EntryHandle, value struct{}) *privateentryHandleSetMap {
	// Grow backing vector if load factor is too high
	if m.Len() >= m.backingVector.Len()*int(upperMapLoadFactor) {
		buckets := newPrivateprivateentryHandleSetMapItemBuckets(m.Len() + 1)
		buckets.AddItemsFromMap(m)
		buckets.AddItem(privateentryHandleSetMapItem{Key: key, Value: value})
		return &privateentryHandleSetMap{backingVector: emptyprivateentryHandleSetMapItemBucketVector.Append(buckets.buckets...), len: buckets.length}
	}

	pos := m.pos(key)
	bucket := m.backingVector.Get(pos)
	if bucket != nil {
		for ix, item := range bucket {
			if item.Key == key {
				// Overwrite existing item
				newBucket := make(privateprivateentryHandleSetMapItemBucket, len(bucket))
				copy(newBucket, bucket)
				newBucket[ix] = privateentryHandleSetMapItem{Key: key, Value: value}
				return &privateentryHandleSetMap{backingVector: m.backingVector.Set(pos, newBucket), len: m.len}
			}
		}

		// Add new item to bucket
		newBucket := make(privateprivateentryHandleSetMapItemBucket, len(bucket), len(bucket)+1)
		copy(newBucket, bucket)
		newBucket = append(newBucket, privateentryHandleSetMapItem{Key: key, Value: value})
		return &privateentryHandleSetMap{backingVector: m.backingVector.Set(pos, newBucket), len: m.len + 1}
	}

	item := privateentryHandleSetMapItem{Key: key, Value: value}
	newBucket := privateprivateentryHandleSetMapItemBucket{item}
	return &privateentryHandleSetMap{backingVector: m.backingVector.Set(pos, newBucket), len: m.len + 1}
}

// Delete returns a new privateentryHandleSetMap without the element identified by key.
func (m *privateentryHandleSetMap) Delete(key *EntryHandle) *privateentryHandleSetMap {
	pos := m.pos(key)
	bucket := m.backingVector.Get(pos)
	if bucket != nil {
		newBucket := make(privateprivateentryHandleSetMapItemBucket, 0)
		for _, item := range bucket {
			if item.Key != key {
				newBucket = append(newBucket, item)
			}
		}

		removedItemCount := len(bucket) - len(newBucket)
		if removedItemCount == 0 {
			return m
		}

		if len(newBucket) == 0 {
			newBucket = nil
		}

		newMap := &privateentryHandleSetMap{backingVector: m.backingVector.Set(pos, newBucket), len: m.len - removedItemCount}
		if newMap.backingVector.Len() > 1 && newMap.Len() < newMap.backingVector.Len()*int(lowerMapLoadFactor) {
			// Shrink backing vector if needed to avoid occupying excessive space
			buckets := newPrivateprivateentryHandleSetMapItemBuckets(newMap.Len())
			buckets.AddItemsFromMap(newMap)
			return &privateentryHandleSetMap{backingVector: emptyprivateentryHandleSetMapItemBucketVector.Append(buckets.buckets...), len: buckets.length}
		}

		return newMap
	}

	return m
}

// Range calls f repeatedly passing it each key and value as argument until either
// all elements have been visited or f returns false.
func (m *privateentryHandleSetMap) Range(f func(*EntryHandle, struct{}) bool) {
	m.backingVector.Range(func(bucket privateprivateentryHandleSetMapItemBucket) bool {
		for _, item := range bucket {
			if !f(item.Key, item.Value) {
				return false
			}
		}
		return true
	})
}

// ToNativeMap returns a native Go map containing all elements of m.
func (m *privateentryHandleSetMap) ToNativeMap() map[*EntryHandle]struct{} {
	result := make(map[*EntryHandle]struct{})
	m.Range(func(key *EntryHandle, value struct{}) bool {
		result[key] = value
		return true
	})

	return result
}

// entryHandleSet is a persistent set
type entryHandleSet struct {
	backingMap *privateentryHandleSetMap
}

// NewentryHandleSet returns a new entryHandleSet containing items.
func NewentryHandleSet(items ...*EntryHandle) *entryHandleSet {
	mapItems := make([]privateentryHandleSetMapItem, 0, len(items))
	var mapValue struct{}
	for _, x := range items {
		mapItems = append(mapItems, privateentryHandleSetMapItem{Key: x, Value: mapValue})
	}

	return &entryHandleSet{backingMap: newprivateentryHandleSetMap(mapItems)}
}

// Add returns a new entryHandleSet containing item.
func (s *entryHandleSet) Add(item *EntryHandle) *entryHandleSet {
	var mapValue struct{}
	return &entryHandleSet{backingMap: s.backingMap.Store(item, mapValue)}
}

// Delete returns a new entryHandleSet without item.
func (s *entryHandleSet) Delete(item *EntryHandle) *entryHandleSet {
	newMap := s.backingMap.Delete(item)
	if newMap == s.backingMap {
		return s
	}

	return &entryHandleSet{backingMap: newMap}
}

// Contains returns true if item is present in s, false otherwise.
func (s *entryHandleSet) Contains(item *EntryHandle) bool {
	_, ok := s.backingMap.Load(item)
	return ok
}

// Range calls f repeatedly passing it each element in s as argument until either
// all elements have been visited or f returns false.
func (s *entryHandleSet) Range(f func(*EntryHandle) bool) {
	s.backingMap.Range(func(k *EntryHandle, _ struct{}) bool {
		return f(k)
	})
}

// IsSubset returns true if all elements in s are present in other, false otherwise.
func (s *entryHandleSet) IsSubset(other *entryHandleSet) bool {
	if other.Len() < s.Len() {
		return false
	}

	isSubset := true
	s.Range(func(item *EntryHandle) bool {
		if !other.Contains(item) {
			isSubset = false
		}

		return isSubset
	})

	return isSubset
}

// IsSuperset returns true if all elements in other are present in s, false otherwise.
func (s *entryHandleSet) IsSuperset(other *entryHandleSet) bool {
	return other.IsSubset(s)
}

// Union returns a new entryHandleSet containing all elements present
// in either s or other.
func (s *entryHandleSet) Union(other *entryHandleSet) *entryHandleSet {
	result := s

	// Simplest possible solution right now. Would probable be more efficient
	// to concatenate two slices of elements from the two sets and create a
	// new set from that slice for many cases.
	other.Range(func(item *EntryHandle) bool {
		result = result.Add(item)
		return true
	})

	return result
}

// Equals returns true if s and other contains the same elements, false otherwise.
func (s *entryHandleSet) Equals(other *entryHandleSet) bool {
	return s.Len() == other.Len() && s.IsSubset(other)
}

func (s *entryHandleSet) difference(other *entryHandleSet) []*EntryHandle {
	items := make([]*EntryHandle, 0)
	s.Range(func(item *EntryHandle) bool {
		if !other.Contains(item) {
			items = append(items, item)
		}

		return true
	})

	return items
}

// Difference returns a new entryHandleSet containing all elements present
// in s but not in other.
func (s *entryHandleSet) Difference(other *entryHandleSet) *entryHandleSet {
	return NewentryHandleSet(s.difference(other)...)
}

// SymmetricDifference returns a new entryHandleSet containing all elements present
// in either s or other but not both.
func (s *entryHandleSet) SymmetricDifference(other *entryHandleSet) *entryHandleSet {
	items := s.difference(other)
	items = append(items, other.difference(s)...)
	return NewentryHandleSet(items...)
}

// Intersection returns a new entryHandleSet containing all elements present in both
// s and other.
func (s *entryHandleSet) Intersection(other *entryHandleSet) *entryHandleSet {
	items := make([]*EntryHandle, 0)
	s.Range(func(item *EntryHandle) bool {
		if other.Contains(item) {
			items = append(items, item)
		}

		return true
	})

	return NewentryHandleSet(items...)
}

// Len returns the number of elements in s.
func (s *entryHandleSet) Len() int {
	return s.backingMap.Len()
}

// ToNativeSlice returns a native Go slice containing all elements of s.
func (s *entryHandleSet) ToNativeSlice() []*EntryHandle {
	items := make([]*EntryHandle, 0, s.Len())
	s.Range(func(item *EntryHandle) bool {
		items = append(items, item)
		return true
	})

	return items
}
