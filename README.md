# Create static blog pages for server, but dynamic from client


During working with github pages, I'm trying to use jekyll, and this is the beginning of my nightmare. 

After running bunch of gem commands followed by the guild(I haven't used ruby before), I got a "jekyll-paginate" (or what ever) not found error. Google told me I should install jekyll-paginate … Shouldn't gem install or bundle takes care dependencies already ? Or what's the meaning of previously running commands ? After installing jekyll-paginate , I got a "Liquid Warning: Liquid syntax error (line 61): Unexpected character & in "site.duoshuo_share && site.duoshuo_username" in /_layouts/post.html" . And someone points this is a jekyll compatible issue. my jekyll shows version 3.5, github pages only supports 3.4, and the blog template I've copied is targeted to 3.0 . Crazy! What I need is just one thing : 

**turn my markdown to browser readable pages**

But from my perspective, almost everything jekyll do can be done by javascript just in client browser. in classic MVC design, your posts is the core model part, jekyll is just a "static view generator" , which with ruby/gem/bundle/liquid confused me very much. so what I'll write is yet another "view generator", but do almost everything in client, instead of generating static pages for server .

How it works:
1. You still put your posts under _posts dir
2. Generator scan all posts to create a multi-level index with all information (such as tags,date,author,title,post file path)
3. Your web site has a front page, javascript read index files, and render your posts

Why javascript instead of jekyll 
1. User can choose different page view style , on the fly, from client. 
2. Easily custom website structure ,without extending jekyll->html stage
3. Dynamic page with javascript much easier than jekyll


How blog index worked: 
* every blog has multi attributes, which can be single value or multi value.
    To improve client response for massive amount blogs, long lists is splited into multi json
* blog can be looked up through any attribute (by its attribute path )
* to reduce file size, custom hash can be used instead of file path
* create a client javascript lib , provides generic client access interface (deal with hash or filepath, auto paging load, client side search, client side history record)


for simple type (personal blog), all information stored in root_index.json

folder structure:

<pre>
    simple type
    webroot
      |_ json_index(folder)
          |_ v0 (folder)
                |_____  root_index.json
</pre>
 
root index: root_index.json under webroot/index/v0 folder, content:
* type (required) : string, only "simple" or "complex" allowed, describe type of index  (complex type will be described next )
* hash (required) : string ,hash function used to map post path into short hash.for massive amount blogs (with verylong path), short hash can reduce index file size,also we can sort index based on number to load exactly index file only.currently supported list: "java32","md5","sha1","sha256","none". java32 is the string.hashCode of java language, simple and create 32bit integer. "none" means no hash used ,any identifier appears later is the actual file path. generator will default try to use weakest(short) hash to create identify, and auto use stronger hash if collide found (finally use actual path instead of hash if all hash has colliding)
* hash_length (optional) : length of hash, only required when hash is not "none". generator will find fewest none-collide chars of hash result(to reduce size of record maximum), note this is count by hex string char number, not count by hash byte.(one byte shows as two chars) . if hash_is not trimmed (eg, 8 for crc32, 32 for md5), hash_length can be omitted

* post_map (optional): required when hash is not "none", object, key is hash of path, value is post path 
```json
"post_map":
{
    "23da1ab0":"post1path",
    "0ac29b2d":"post2path"
}
```
    note: hash is stored as hex string, and only first hash_length number of chars 
    (not bytes, which represent as 2 chars)shows. for example full java32 hash 
    value 3d5a8912, and hash_length is 3, only "3d5" is preserved.

* attributes: object, key is the attribute name of post, value is the detail information of this attribute
```json
 "tag":{
 "is_multi":true  
 }
```
   is_multi : if one post can contain multi value of this attribute.(not means if their value is unique, for example , two posts can have same title, but one post has only one title ,so title still set is_multi=false)
 
attributes example:
```json
"attributes":
{
  "tag":{
    "is_multi":true
  },
  "date":{
    "is_multi":false
  }
}
```

* attributes_index: object, key is the attribute name , value is "attribute value:[ ]" 
                           elements in array is always sorted (for simple type)
```json
"attributes_index":{
     "title":
     {
         "title1":["hash-or-postpath1","hash-or-postpath2"],
         "title2":["hash-or-postpath2"]
     },
     "date":
     {
         "2017-12-12 hh:mm:ss":["hash-or-postpath1","hash-or-postpath2"],
         "2017-12-13 hh:mm:ss":["hash-or-postpath3","hash-or-postpath4"]
     },
     "tag":
     {
       "tagA":["hash-or-postpath1","hash-or-postpath2"],
       "tagB":["hash-or-postpath3"]
     }
   }
```

complete example of root_index.json for simple type:
```json
{
  "type":"simple",
  "hash":"md5",
  "hash_length":3,
  "use_post_map":"true",
  "post_map":
      {
         "a52":"_posts/post1.md",
         "bbc":"_posts/post2.md",
         "189":"_posts/post3.md",
      },
  "attributes":
    { 
         "author":{"is_multi":false},
         "date":{"is_multi":false},
         "tag":{"is_multi":true},
         "title":{"is_multi":false}
    },
    "attributes_index":
    {
      "author":{
        "me":["189","a52"],
        "you":["bbc"]
      },
      "date":{
      "2017-04-02 15:23:03":["bbc"],
      "2017-05-12 08:01:69":["189"],
      "2017-07-21 02:41:15":["a52"],
      },
      "tag":{
        "c++":["bbc"],
        "java":["189","bbc"],
        "python":["189","a52"]
      },
      "title":{
      "call python in java":["189"],
      "JIT":["a52"],
      "JNI in c++":["bbc"]
      }
    }
}
```


complex type folder structure:
<pre>
    webroot
      |_ json_index(folder)
          |_ v0 (folder)
                |_____  root_index.json    
                |__post_map  folder to store post_map jsons
                |     |___0.json  --> post_map member moved to these jsons
                |     |___1.json
                |__attributes
                |     |___0.json   --> attributes member moved to these jsons
                |     |___1.json
                |__attributes_index(folder)  --> sub index of every attribute
                      |___single
                      |     |_____ att0_0.json       --> sub index of is_multi=false attribute
                      |     |_____ att0_1.json       --> sub index of is_multi=false attribute
                      |     |_____ att1            --> folder of is_multi=true attribute
                      |___multi
                            |___name0_0.json --> sub attribute index
                            |___name0_1.json --> sub attribute index
                            |___name1_0.json --> sub attribute index
                            |___name1_1.json --> sub attribute index
</pre>


for type "complex" type (lots of blogs), root_index.json contains:
* type: "complex"
* total: integer ,total number of all posts
* hash : same as simple
* hash_length : same as simple
* post_map (optional): object contains page and order info. required when hash is not "none"(but can be omitted if all member is default)
  - pages (optional): integer number of total post_map pages (for example 100 record per json, 178 record totally, will result pages=2), no such member means pages=1. sub page always have less equal than 256 in one folder, for example pages <=256, they all appears under post_map folder, but pages = 65000 , will create folder structure as follwing : 
<pre>
    post_map   --> root folder to store post_map sub index
       |__ 0   --> spare folder  
       |   |__0.json
       |   |__1.json
       |   |__...
       |   |__255.json
       |__ 1   --> spare folder
       |   |__256.json
       |   |__257.json
       |   |__...json
       |   |__511.json
       |__ 2   --> spare folder
</pre>
if pages >= 256*256, it will create 3 levels folder structure (less equal than 256 files in one folder , less equal than 256 folder under parent folder)
  - sorted (optional): how record in post_map ordered ,  can be "alpha" only (order by string),  if no such member, means unordered 
  - order (optional): if 'sorted' provided , order describe "desc" or "asc" , if no such member, default to "asc"
  - range (optional): required when 'sorted' provided . it's a array of 'pages'+1 elements, first element is the first record of page0 file, following 'pages' elements is the last key of every paged file records. this field is provided for fast lookup into sub-json. 
       for example you have 3 pages ordered post_map which have range 
       ["000","100","200","300"] which means your 3 pages range is 000 ~ 100, 101 ~ 200, 201 ~ 300
       and want to load post of hash="132" , you can binary search in range array to find which page this post_map exactly in , and load that json directly, without iterate all sub pages
  example:
```json
{
  "post_map":
  {
    "pages":2,
    "sorted":"alpha",
    "order":"asc",
    "range":["893","f2a","fff"]
  }
}
```
note: when every member of post_map is default, you can completely omit post_map member 

post_map/n/…/n.json : splited post_map json file, example:
```json
{
  "08273cd":"_posts/2015/02/11/my-first-blog.md",
  "a3914ef":"_posts/2016/09/12/your-second-blog.md"
}
```
* attributes (optional): object, describe index of attributes splitted json 

example:
```json
{
  "attributes":
  {
    "total":25,
    "pages":2,
    "sorted":"alpha",
    "order":"desc",
    "range":["author","date","title"]
  }
}
```
   object member of attributes almost same as post_map 
* attributes object in simple type moved to splited attributes/…/0.json ~ attributes/…/n.json paged json files(also spared with 256 limit), with following additional fields:
  - total (required): how many attributes we have
  - pages (optional) : pages number of this attribute sub index , default to 1 if omitted
  - safe_name (optional): required when attribute name have when attribute name have special chars (space,slash,backslash,star or any others not safe to be part of path). for complex type, every attribute index moved to its named file, better not to have attribute name with special chars, so this member will be the safe replace of original name (with special chars replaced by under score '_', generator should be response to avoid two safe_name clash each other). if omitted, safe_name should be same as original name
  - sorted : optional: tell if sub index of this attribute is sorted in splited files 
                                  value can be "alpha" , "numeric",  not specified means not sorted
  - range (optinal) : required when 'sorted' and pages >1, range member is a array of "pages"+1 elements, describe first record and every last element of pages end. for example tags attribute have 4 pages, sub index file tags tags_0.json stored  tag00 tag01 tag02... tag20 informations in order, and first stored tag is tag20 (tag20 may still appears in next sub index file). note: range of attribute can have duplicate ranges (which post_map hash is always unique so never happen), for example you may have many elements of 'tags35', across multi json, which results a range like 
```json
 "range":["tag00","tag35","tag35","tag35","tag35","tag42"]
```
because paged json is split by record number (which can be always balance), not by "key" (which can be imbalance)

example: 
attributes/0.json
```json
{
    "tag":{
      "is_multi":true,
      "total":38,
      "pages":4,
      "sorted":"alpha",   
      "order":"asc",
        "range":["tag20","tag35","tag48","tag62","tag90"] 
        },
    "date":{
      "total":41,
      "is_multi":false,
      "pages":2,
      "sorted":true,
      "order":"desc",
      "range":["2017-06-01 10:47:12","2017-04-12 17:10:01"]
      }
}
```
attributes/1.json
```json
        {
              "catalog":{
              "total":5,
              "is_multi":false,
              "pages":3,
              "sorted":"alpha",
              "order":"desc",
              "range":["news","IT"]
              },
              "skill points":{
              "safe_name":"skill_points",
              "total":5,
              "is_multi":false,
              "pages":2,
              "sorted":"numeric",
              "order":"desc",
              "range":[120,100,50]
              }
        }
```

subindex file : attributes_index object in simple type json moved to splited json under attributes_index folder, 
is_multi=false will create 1 level json
for example title attribute, subtitle attribute is_multi=false, so they will create folder structure as 
<pre>
      attributes_index
          |___ title 
          |      |_0.json
          |      |_1.json
          |___ subtitle
                 |_0.json  
                 |_1.json  
</pre>
or if title pages and subtitle pages all greator than 256:
<pre>
      attributes_index
          |___ title
          |      |__0  ==> folder
          |      |  |__0.json
          |      |  |__…json
          |      |  |__255.json
          |      |
          |      |__1  ==> folder
          |         |__256.json
          |____subtitle
                 |__0 ==> folder
                 |  |__0.json
                 |  |__...json
                 |  |__255.json
                 |
                 |__1 ==> folder
                    |_256.json

</pre>

for is_multi=true attributes , such as tags, it will create 2 levels folder index folder structure as
<pre>
      attributes_index
          |___ tags  (folder)
                |___ c++
                |      |__0.json
                |      |__1.json
                |___ php
                       |__0.json
                       |__1.json
</pre>
note : when pages > 256, they follow same rules as 1 level attributes (spare by folder)

note : if is_multi=true attribute have more than 256 values, it also been paged into spared folder ,split by alpha, for example
<pre>
      attributes_index
          |___ tags  (folder)
                |___ 0  
                |    |___c++
                |    |     |__0.json
                |    |     |__1.json
                |    |___ c
                |    |    |__0.json
                |    |    |__1.json
                |    |___ C#
                |         |__0.json
                |         |__1.json
                |
                |
                |___ 1  
                     |____php
                           |__0.json
                           |__1.json
</pre>
