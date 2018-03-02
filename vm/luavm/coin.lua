-- require module
local ZIP = require("ZIP")

-- 合约创建时会被调用一次，之后就不会被调用
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
