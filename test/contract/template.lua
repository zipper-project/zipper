-- require module
local ZIP = require("ZIP")

-- init
function Init(args)
    print("init...", args)
    return true
end

-- invoke when exec transaction
function Invoke(func, args)
    print("in Invoke", func, args)
    return true
end

-- query
function Query(args)
    return "query ok"
end
