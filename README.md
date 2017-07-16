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


How blog index (which is the most important) constructed:
    every blog is identified by hash (configurable) of its path (relative to _posts directory) and name (md5 should be enough for your post names)
    every blog has multi attributes, which can be single value or multi value.
    To improve client response for massive amount blogs, long lists is splited into multi json

folder structure:

<pre>
    simple type
    webroot
      |_ json_index(folder)
          |_ v0 (folder)
                |_____  root_index.json
</pre>
 
root index: root_index.json under webroot/index/v0 folder, content:
* type : string, only "simple" or "complex" allowed, describe type of index  (complex type will be described next )
* hash : string ,hash function used to map post path into short hash.for massive amount blogs (with verylong path), short hash can reduce index file size,also we can sort index based on number to load exactly map only.currently supported list: "java32","md5","sha1". java32 is the string.hashCode of java language, simple and create 32bit integer
* hash_length : length of hash. generator will find fewest none-collide chars of hash result, and auto use stronger hash if collide found (java32->md5->sha1). this is count by hex string char number, not count by hash byte.


for type:"simple" content, all information stored in root_index.json

* use_post_map: bool true=using .post_map to map hash<=>post path , 
                              all post reference use hash instead of post path
                              false=not using hash, all value for post is 
                              file path, not hash
* attributes: object, key is the attribute name, value is a information shows as following
```json
 "tag":{
 "is_multi":true  
 }
```
  is_multi : does one post can contain multi value of this attribute.(not means if their value is unique, for example , two posts can have same title, but one post has only one title ,so title still set is_multi=true )
 
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


* post_map: object, key is hash of path, value is post path (optional only when .use_post_map=true)
```json
"post_map":
{
    "23da1ab0":"post1path",
    "0ac29b2d":"post2path"
}
```
    note: hash is stored as hex string, and only first hash_length number of chars (not bytes, which represent as 2 chars)shows. for example full java32 hash value 3d5a8912, and hash_length is 3, only "3d5" is preserved.

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
       "tagA":["post1path","post2path"]
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
      "JNI in c++":["bbc"],
      "call python in java":["189"],
      "JIT":["a52"]
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
                |_____  post_map_0.json  --> post_map member moved to these jsons
                |_____  post_map_1.json
                |_____  attributes_0.json   --> attributes member moved to these jsons
                |_____  attributes_1.json
                |_____  attributes_index(folder)  --> sub index of every attribute
                           |_____ att0name_0.json
                           |_____ att0name_1.json
                           |_____ att1name_0.json
                           |_____ att2name_0.json
</pre>


for type "complex" content:
- root_index.json:
  * no use_post_map member, post_map is always true
  * post_map: object integer number of total post_map pages (for example 100 record per json, 178 record totally, will result post_map_pages=2)
```json
{
  "post_map":
  {
    "pages":2,
    "sorted":"alpha",
    "order":"asc",
    "range":["893","f2a"]
  }
}
```
  * attributes: integer number of total attribute pages 

* attributes array in simple type moved to splited attributes_0.json ~ attributes_n.json paged json files, with following additional fields:
  - pages : pages number of this attribute sub index
  - sorted : optional: tell if sub index of this attribute is sorted in splited files 
                                  value can be "alpha" , "numeric",  not specified means not sorted
  - range : array if have sorted , range member is a array of "pages" elements, describe the last element of every pages end. for example tags attribute have 5 pages, sub index file tags tags_0.json stored  tag00 tag01 tag02... tag20 informations in order, and last stored tag is tag20 (tag20 may still appears in next sub index file)
example: 
        attributes_0.json
```json
{
"attributes":{
    "tag":{
      "is_multi":true,
      "pages":5,
      "sorted":"alpha",   
      "order":"asc",
        "range":["tag20","tag35","tag48","tag62","tag90"] 
        },
    "date":{
      "is_multi":false,
      "pages":2,
      "sorted":true,
      "order":"desc",
      "range":["2017-06-01 10:47:12","2017-04-12 17:10:01"]
      }
    }
}
```
        attributes_1.json
```json
        {
              "name":{"catalog","pages":3},
              "name":{"author","pages":1}
        ]
        }
```
        one "pages" member added to element of attributes array ,
        specified pages of this attribute sub-index (following)
         note: top level element is still a json object



in attribute array of unique attributes info of posts. info contains: 
* post map index:
post_map_n.json: 
```json
{
  "post_map":
  {
  "08273cd":"_posts/2015/02/11/my-first-blog.md",
  "a3914ef":"_posts/2016/09/12/your-second-blog.md"
  }
}
```
subindex file

