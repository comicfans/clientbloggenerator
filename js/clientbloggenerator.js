"strict"

function LoadIndex(indexUrl,callback){

    $.getJson(indexUrl).then(function(data ,status, xhr){
    });

}

(function(exports){

    exports.CurrentDefaultIndexUrl= function(){
        return location.protocol+'//'+location.hostname+(location.port ? ':'+location.port: '');
    };


    iteratePosts=function(data, postCallback){
    }
    
    parseIndex=function(data){

        data.prototype.IteratePosts=function(){
        }

        if (data.type === 'simple') {
            return;
        }

    };

    exports.Load=function(url,loadedFunc){
            
        $.getJSON(url).done(function(data){
            parseIndex(data);
            loadedFunc(data);
        }).fail(function(){loadedFunc(null);});
    }


    exports.Default = function(loadedFunc){
        exports.Load(exports.CurrentDefaultIndexUrl(),loadedFunc)
    }
})(typeof exports === 'undefined'?this['clientbloggenerator']={}:exports);
