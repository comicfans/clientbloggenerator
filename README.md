Create static blog pages for server, but dynamic from client


During working with github pages, I'm trying to use jekyll, and this is the beginning of my nightmare. 

After running bunch of gem commands followed by the guild(I haven't used ruby before), I got a "jekyll-paginate" (or what ever) not found error. Google told me I should install jekyll-paginate … Shouldn't gem install or bundle takes care dependencies already ? Or what's the meaning of previously running commands ? After installing jekyll-paginate , I got a "Liquid Warning: Liquid syntax error (line 61): Unexpected character & in "site.duoshuo_share && site.duoshuo_username" in /_layouts/post.html" . And someone points this is a jekyll compatible issue. my jekyll shows version 3.5, github pages only supports 3.4, and the blog template I've copied is targeted to 3.0 . Crazy! What I need is just one thing : 

turn my markdown to browser readable pages 

But from my perspective, almost everything jekyll do can be done by javascript just in client browser. in classic MVC design, your posts is the core model part, jekyll is just a "static view generator" , which with ruby/gem/bundle/liquid confused me very much. so what I'll write is yet another "view generator", but do almost everything in client, instead of generating static pages for server .

How it works:
    1. You still put your posts under _posts dir
    2. Generator scan all posts to create a multi-level index with all information (such as tags,date,author,title,post file path)
    3. Your web site has a front page, javascript read index files, and render your posts

Why javascript instead of jekyll 
    1. User can choose different page view style , on the fly, from client. 
  



How blog index (which is the most important) constructed:
    every blog is identified by md5 of its path (relative to _posts directory) and name (md5 should be enough for your post names)
    every blog has multi attributes, which can be single value or multi value.
    To improve client response for massive amount blogs, long lists is splited into multi json

    folder structure:
    webroot
      |_ json_index(folder)
          |_ v0 (folder)
                |_____  root_index.json
                |_____  post_map_0.json
                |_____  post_map_1.json
                |_____  attributes_index(folder) 
                           |_____ att0_0.json
                           |_____ att0_1.json
                           |_____ att2_0.json
                           |_____ att3_0.json

    root index:
      root_index.json under webroot/index/v0 folder
      content:
        .hash : string ,hash function used to map post path into short hash 
              if "none" used, post_map is ignored completly, every entry saved is just the post path
              for massive amount blogs (with verylong path), short hash can reduce index file size,
              also we can sort index based on number to load exactly map only

        .post_map_number int  number of total post_map pages (for example 100 record per json, 178 record totally, number)
        .attributes : []  array of unique attributes info of posts. info contains: 
            .name : attribute name (no support space inside attribute name, but you can choose to display any name from client)
            .page_number : int number of total paged 
            .is_multi :  bool value should this attribute have multi values (for example tags)
            

    post map index:
      post_map_n.json: 
        .key is md5 of post path  
        .value is the post path
        for example
        {
          "a4dadca134123769721abedf92021234":"2015/01/22/my-first-blog.md",
          {},
          {}
        }


       
       
