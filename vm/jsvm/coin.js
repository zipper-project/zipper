/***** JS example *****/

// only exec once when deploy contract
function Init(args) {
    console.log("init", args)
    return true;
}

// to trigger invoke when exec contract
function Invoke(func, args) {
    console.log("invoke", func, args)
    return true;
}

// to trigger query when exec query
function Query(args) {
    console.log("call Query", args);
    return "query ok"
}
