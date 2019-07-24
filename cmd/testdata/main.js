
// Duktape.modSearch = function (id, require, exports, module) {
    
//     var res;

//     print('loading module:', id);

//     res = readFile('/modules/' + id + '.js');
//     if (typeof res === 'string') {
//         return res;
//     }

//     throw new Error('module not found: ' + id);
// }

var sec = require("./second.js");

// print(JSON.stringify(sec));

// print("hello");
// print(sec)

// print(JSON.stringify(sec))

var myval = { 
    something: sec.coolStoryBro.fuck,
    could: "be",
    happening: 1231234,
}

myval