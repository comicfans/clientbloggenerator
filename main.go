package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"github.com/gernest/front"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/syncmap"
	"hash"
	"hash/crc32"
	"hash/crc64"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	log = logrus.New()
)

const (
	STRING  = "string"
	NUMERIC = "numeric"
	DATE    = "date"
)

var GUESS_TIME_LAYOUT []string = []string{"1987-11-28 13:22:48 +0800", "1989-1-3 0:2:23", "2014-4-27 13:20", "2003-9-28 11", "2010-9-3"}

type AttributeInfo struct {
	is_multi    bool
	is_numeric  bool
	is_date     bool
	value_posts map[string][]*PostInfo
	numeric_sum int64
}

type RootIndex struct {
	wg              sync.WaitGroup
	post_infos      syncmap.Map
	raw_name_length uint64
	post_number     uint64
	hash            string
	hash_length     int
	post_map        map[string]string
	attributes_map  map[string]*AttributeInfo
}

type PostInfo struct {
	hash_string string
	attributes  map[string]interface{}
}

const (
	CRC32    = iota
	CRC64    = iota
	MD5      = iota
	SHA1     = iota
	SHA224   = iota
	SHA256   = iota
	SHA384   = iota
	SHA512   = iota
	MAX_HASH = iota
)

var HASH_HEX_LENGTH = make([]int, MAX_HASH)

var HASH_FUNC = make([](func() hash.Hash), MAX_HASH)

var HASH_NAME = make([]string, MAX_HASH)

var CRC64_TABLE = crc64.MakeTable(crc64.ISO)

func init() {
	HASH_HEX_LENGTH[CRC32] = 8
	HASH_HEX_LENGTH[CRC64] = 16
	HASH_HEX_LENGTH[MD5] = 32
	HASH_HEX_LENGTH[SHA1] = 40
	HASH_HEX_LENGTH[SHA224] = 56
	HASH_HEX_LENGTH[SHA256] = 64
	HASH_HEX_LENGTH[SHA384] = 96
	HASH_HEX_LENGTH[SHA512] = 128

	HASH_FUNC[CRC32] = func() hash.Hash { return crc32.NewIEEE() }
	HASH_FUNC[CRC64] = func() hash.Hash { return crc64.New(CRC64_TABLE) }
	HASH_FUNC[MD5] = md5.New
	HASH_FUNC[SHA1] = sha1.New
	HASH_FUNC[SHA224] = sha256.New224
	HASH_FUNC[SHA256] = sha256.New
	HASH_FUNC[SHA384] = sha512.New384
	HASH_FUNC[SHA512] = sha512.New

	HASH_NAME[CRC32] = "crc32"
	HASH_NAME[CRC64] = "crc64"
	HASH_NAME[MD5] = "md5"
	HASH_NAME[SHA1] = "sha1"
	HASH_NAME[SHA224] = "sha224"
	HASH_NAME[SHA256] = "sha256"
	HASH_NAME[SHA384] = "sha384"
	HASH_NAME[SHA512] = "sha512"
}

func shouldProcess(path string) bool {

	if strings.HasSuffix(path, ".swp") {
		return false
	}
	if strings.Contains(path, "_post") {
		return true
	}
	return false
}

func (index *RootIndex) AddPostInfo(postPath string, postInfo *PostInfo) {

	atomic.AddUint64(&index.raw_name_length, uint64(len(postPath)))
	atomic.AddUint64(&index.post_number, 1)

	index.post_infos.Store(postPath, postInfo)

}

func (index *RootIndex) Parse(path string) {

	r, err := os.Open(path)
	if err != nil {
		log.Errorf("error reading file %s", err)
		return
	}

	defer r.Close()

	m := front.NewMatter()

	m.Handle("---", front.YAMLHandler)

	f, _, err := m.Parse(r)
	if err != nil {
		log.Errorf("error parseing file %s front header ", path)
		return
	}

	var tothrow []string
	for key, value := range f {
		if _, ismap := value.(map[string]interface{}); ismap {
			tothrow = append(tothrow, key)
			continue
		}
	}

	for _, key := range tothrow {
		log.Warnf("YAML tag '%s' value type is not supported(a map)", key)
		delete(f, key)
	}

	for attr, _ := range f {
		log.Debugf("attr %s", attr)
	}

	postInfo := &PostInfo{
		attributes: f,
	}

	index.AddPostInfo(path, postInfo)
}

func (index *RootIndex) Visit(path string, f os.FileInfo, err error) error {

	if f.IsDir() {
		return nil
	}

	if !shouldProcess(path) {
		return nil
	}

	index.wg.Add(1)
	go func() {
		log.Infof("start processing %s", path)
		index.Parse(path)
		index.wg.Done()
		log.Infof("done processing %s", path)
	}()

	return nil
}

func CollectPosts() *RootIndex {

	rootIndex := &RootIndex{
		raw_name_length: 0,
		post_number:     0,
		hash:            "none",
		hash_length:     0,
		post_map:        make(map[string]string),
		attributes_map:  make(map[string]*AttributeInfo),
	}

	cwd, err := os.Getwd()

	if err != nil {
		log.Fatal("can not determine cwd")
		return nil
	}

	filepath.Walk(cwd, rootIndex.Visit)

	rootIndex.wg.Wait()

	return rootIndex
}

func TestColliding(hashMap *syncmap.Map, prevTryLength, hashMaxLength int) (int, map[string]string) {

	for l := prevTryLength + 1; l <= hashMaxLength; l++ {

		testMap := make(map[string]string)

		colliding := false
		hashMap.Range(func(key, value interface{}) bool {
			stringKey := hex.EncodeToString([]byte(key.(string)))[:l]

			log.Debugf("key %v", stringKey)

			if _, has := testMap[stringKey]; has {
				colliding = true
				return false
			}

			testMap[stringKey] = value.(string)

			return true
		})

		if !colliding {

			log.Infof("length %v is enough", l)
			return l, testMap
		}

		log.Debugf("length %v collding", l)

	}

	return hashMaxLength, nil
}

func (rootIndex *RootIndex) FindShortestHash() {

	for i := 0; i < MAX_HASH; i++ {

		prevTryLength := 0
		if i != 0 {
			prevTryLength = HASH_HEX_LENGTH[i-1]
		}

		if rootIndex.raw_name_length < uint64(prevTryLength)*rootIndex.post_number {
			log.Info("post name length < min hash name length, use post name directly")
			rootIndex.hash = "none"
			return
		}

		var tryMap syncmap.Map

		var wg sync.WaitGroup

		colliding := false
		rootIndex.post_infos.Range(func(postPath, postInfo interface{}) bool {

			wg.Add(1)
			go func() {

				defer wg.Done()

				h := HASH_FUNC[i]()

				h.Write([]byte(postPath.(string)))
				sum := h.Sum(nil)

				postInfo.(*PostInfo).hash_string = hex.EncodeToString(sum)
				log.Debugf("%v hash is %v", postPath, postInfo.(*PostInfo).hash_string)

				_, loaded := tryMap.LoadOrStore(string(sum), postPath)

				if loaded {
					colliding = true
				}
			}()

			return !colliding
		})

		wg.Wait()

		if colliding {
			log.Debug(HASH_NAME[i], " has collising, skip")
			continue
		}

		log.Debug(HASH_NAME[i], " is enough, try shortest length")

		rootIndex.hash = HASH_NAME[i]
		hashMaxLength := HASH_HEX_LENGTH[i]
		rootIndex.hash_length, rootIndex.post_map = TestColliding(&tryMap, prevTryLength, hashMaxLength)

		//this is shortest hash
		log.Infof("choose %s as hash ,hash length %d", HASH_NAME[i], rootIndex.hash_length)
		if rootIndex.hash_length == hashMaxLength {
			// same as max length , can be omitted
			rootIndex.hash_length = 0
		}
		return
	}

	log.Info("all hash hash colliding, using post name directly")
	rootIndex.hash = "none"

}

type DateStrings []string

func (data DateStrings) Len() int {
	return len(data)
}

func (data DateStrings) Swap(i, j int) {
	data[i], data[j] = data[j], data[i]
}

func (data DateStrings) Less(i, j int) bool {

	dI, _ := guessDateValue(data[i])
	dJ, _ := guessDateValue(data[j])

	return (*dI).Before(*dJ)

}

type NumericStrings []string

func (data NumericStrings) Less(i, j int) bool {
	nI, _ := strconv.Atoi(data[i])
	nJ, _ := strconv.Atoi(data[j])
	return nI < nJ
}

func (data NumericStrings) Len() int {
	return len(data)
}

func (data NumericStrings) Swap(i, j int) {
	data[i], data[j] = data[j], data[i]
}

type Attributes RootIndex

func (attributes *Attributes) MarshalText() (text []byte, err error) {

	var orderedList []string
	for key, _ := range attributes.attributes_map {
		orderedList = append(orderedList, key)
	}

	sort.Strings(orderedList)

	var buf bytes.Buffer
	{
		buf.WriteByte('{')
		defer buf.WriteByte('}')

		for i, key := range orderedList {
			if i != len(orderedList)-1 {
				buf.WriteByte(',')
			}

			str, _ := json.Marshal(key)
			buf.Write(str)
			buf.WriteString(":{")
			defer buf.WriteByte('}')

			attrInfo := attributes.attributes_map[key]

			var sortedAttrs []string
			for key, _ := range attrInfo.value_posts {
				sortedAttrs = append(sortedAttrs, key)
			}

			var valueType string
			if attrInfo.is_date {
				sort.Sort(DateStrings(sortedAttrs))
				valueType = DATE
			} else if attrInfo.is_numeric {
				sort.Sort(NumericStrings(sortedAttrs))
				valueType = NUMERIC
			} else {
				sort.Strings(sortedAttrs)
				valueType = STRING
			}

			buf.WriteString(`"value_type":`)
			buf.WriteString(valueType)
			buf.WriteByte(",")

			if attrInfo.is_numeric {
				jsonStr, _ := json.Marshal("numeric_sum")
				buf.Write(jsonStr)
				buf.WriteByte(':')
				jsonStr, _ = json.Marshal(attrInfo.numeric_sum)

				buf.Write([]byte(jsonStr))
			}

		}
	}

	return buf.Bytes(), nil
}

type AttributesIndex RootIndex

func (attributesIndex *AttributesIndex) MarshalText() (text []byte, err error) {

	return nil, nil
}

func (rootIndex *RootIndex) GenerateIndexJson(typeIsSimple bool) {

	jsonObject := make(map[string]interface{})

	jsonObject["type"] = "simple"

	usePostMap := false
	if rootIndex.hash != "none" {
		usePostMap = true
		jsonObject["use_post_map"] = usePostMap
		jsonObject["hash"] = rootIndex.hash
	}
	if rootIndex.hash_length != 0 {
		jsonObject["hash_length"] = rootIndex.hash_length
	}

	if usePostMap {
		jsonObject["post_map"] = rootIndex.post_map
	}

	jsonObject["attrbiutes"] = (*Attributes)(rootIndex)
	jsonObject["attrbiutes_index"] = (*AttributesIndex)(rootIndex)

	indexDir := filepath.Join("json_index", "v0")

	err := os.MkdirAll(indexDir, os.ModePerm)
	if err != nil {
		log.Fatalf("can not create index dir %s", indexDir)
		return
	}

	indexFile := filepath.Join(indexDir, "root_index.json")
	writer, err := os.Create(indexFile)
	if err != nil {

		log.Fatalf("can not create json file %v,%v", indexFile, err)
		return
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", " ")

	err = encoder.Encode(jsonObject)
	if err != nil {
		log.Fatalf("error writing json file %v, %v", indexFile, err)
		return
	}

	log.Infof("create index file %v", indexFile)

}

func guessDateValue(str string) (t *time.Time, ok bool) {

	for _, guess := range GUESS_TIME_LAYOUT {
		t, err := time.Parse(guess, str)
		if err == nil {
			return &t, true
		}
	}
	return nil, false
}

func GuessValueType(attrInfo *AttributeInfo, attrValue string) {

	integer, err := strconv.Atoi(attrValue)
	if err != nil {
		attrInfo.is_numeric = false
		attrInfo.numeric_sum = -1
	} else if attrInfo.is_numeric {
		attrInfo.numeric_sum = attrInfo.numeric_sum + int64(integer)
	}

	if _, noLayoutMatch := guessDateValue(attrValue); noLayoutMatch {
		attrInfo.is_date = false
	}
}

func AddAttrPost(m *map[string]*AttributeInfo, attrName string, attrValues []string, postInfo *PostInfo, hash_length int) {

	attrInfo, ok := (*m)[attrName]
	if !ok {
		attrInfo = &AttributeInfo{
			is_multi:    false,
			is_numeric:  true,
			is_date:     true,
			value_posts: make(map[string][]*PostInfo),
			numeric_sum: 0,
		}
		(*m)[attrName] = attrInfo
	}

	if len(attrValues) > 1 {
		attrInfo.is_multi = true
	}

	for _, attrValue := range attrValues {

		GuessValueType(attrInfo, attrValue)

		list, _ := attrInfo.value_posts[attrValue]

		attrInfo.value_posts[attrValue] = append(list, postInfo)
	}

}

func (rootIndex *RootIndex) GenerateReverseMap() {

	rootIndex.post_infos.Range(func(key, value interface{}) bool {

		postInfo := value.(*PostInfo)

		for attrName, value := range postInfo.attributes {

			var attrValues []string

			if v, ok := value.(string); ok {
				attrValues = append(attrValues, v)
			} else if v, ok := value.([]interface{}); ok {
				for _, i := range v {
					attrValues = append(attrValues, i.(string))
				}
			} else {
				log.Fatalf("bad logic ,incorrect type %T", value)
			}

			AddAttrPost(&rootIndex.attributes_map, attrName, attrValues, postInfo, rootIndex.hash_length)
		}
		return true
	})
}

func main() {

	log.SetLevel(logrus.DebugLevel)

	rootIndex := CollectPosts()

	rootIndex.FindShortestHash()

	rootIndex.GenerateReverseMap()

	rootIndex.GenerateIndexJson(true)

}
