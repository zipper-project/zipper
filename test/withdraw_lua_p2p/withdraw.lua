local L0 = require("ZIP")

local CName = "withdraw"
local Scale = 1
string.split = function(s, p)
    local rt= {}
    string.gsub(s, '[^'..p..']+', function(w) table.insert(rt, w) end )
    return rt
end
table.size = function(tb) 
    local cnt = 0
    for k, v in pairs(tb) do 
        cnt = cnt + 1
    end
    return cnt
end

function Init(args)
    -- info
    local txInfo = L0.TxInfo()
    local sender = txInfo["Sender"]
    print(">>> init INFO:", txInfo["Hash"], sender)


    local str = ""
    for k, v in pairs(args) do 
        str = str .. v .. ","
    end
    print("INFO:" .. CName .. " L0Init(" .. string.sub(str, 0, -2) .. ")")

    -- validate
    if(table.size(args) ~= 2)
    then
        print("ERR :" .. CName ..  " L0Init --- wrong args number", table.size(args))
        return false
    end
    
    -- execute
    L0.PutState("version", 0)
    print("INFO:" .. CName ..  " L0Init --- system account " .. args[0])
    L0.PutState("account_system", args[0])
    print("INFO:" .. CName ..  " L0Init --- fee account " .. args[1])
    L0.PutState("account_fee", args[1])
    return true
end

function Invoke(func, args)
    -- info
    print('========================================')
    local str = ""
    for k, v in pairs(args) do 
        str = str .. v .. ","
    end
    local txInfo = L0.TxInfo()
    local sender = txInfo["Sender"]
    print(">>> INFO:" .. CName ..  " L0Invoke(" .. func .. "," .. string.sub(str, 0, -2) .. ")", txInfo["Hash"], sender)

    -- execute
    if("launch" == func) then
        return launch(args)
    elseif("cancel" == func) then
        return cancel(args)
    elseif("succeed" == func) then
        return succeed(args)
    elseif("fail" == func) then
        return fail(args)
    else
        print("ERR :" .. CName ..  " L0Invoke --- function not support", func)
        return false
    end
    return true
end

function Query(args)
    -- print info
    local str = ""
    for k, v in pairs(args) do 
        str = str .. v .. ","
    end
    print("INFO:" .. CName ..  " L0Query(" .. string.sub(str, 0, -2) .. ")")
    -- validate
    if(table.size(args) ~= 1)
    then
        print("ERR :" .. CName ..  " L0Query --- wrong args number", table.size(args))
        return false
    end
    -- execute
    local withdrawID = "withdraw_"..args[0]
    local withdrawInfo = L0.GetState(withdrawID)
    if (not withdrawInfo)
    then
        return args[0] .. " not found "
    end
    local tb = string.split(withdrawInfo, "&")
    local addr = tb[1]
    local assetID = tonumber(tb[2])
    local amount = tonumber(tb[3])/Scale
    return args[0] .. " addr:" .. addr .. " , asset:" .. assetID .. " , amount:" .. amount
end

function launch(args) 
    -- validate
    if(table.size(args) ~= 1)
    then
        print("ERR :" .. CName ..  " launch --- wrong args number", table.size(args))
        return false
    end

    -- execute 
    local withdrawID = "withdraw_"..args[0]
    ----[[
    if (L0.GetState(withdrawID))
    then
        print("ERR :" .. CName ..  " launch --- withdrawID alreay exist", args[0])
        return false
    end
    local txInfo = L0.TxInfo()
    local sender = txInfo["Sender"]
    local assetID = txInfo["AssetID"]
    local amount = txInfo["Amount"]
    if (type(sender) ~= "string")
    then
        print("ERR :" .. CName ..  " launch --- wrong sender", sender)
        return false
    end
    if (type(assetID) ~= "number" or assetID < 0)
    then
        print("ERR :" .. CName ..  " launch --- wrong assetID", assetID)
        return false
    end
    if (type(amount) ~= "number" or amount <= 0)
    then
        print("ERR :" .. CName ..  " launch --- wrong amount", amount)
        return false
    end
    L0.PutState(withdrawID, sender.."&"..assetID.."&"..amount)
    print("INFO:" .. CName ..  " launch ---", withdrawID, sender, assetID, amount)
    --]]--
    return true
end

function cancel(args)
    -- validate
    if(table.size(args) ~= 1)
    then
        print("ERR :" .. CName ..  " cancel --- wrong args number", table.size(args))
        return false
    end
    -- execute
    local withdrawID = "withdraw_"..args[0]
    ----[[
    local withdrawInfo = L0.GetState(withdrawID)
    if (not withdrawInfo) 
    then
        print("ERR :" .. CName ..  " cancel --- withdrawID not exist", args[0])
        return false
    end
    local txInfo = L0.TxInfo()
    local sender = txInfo["Sender"]
    local amount = txInfo["Amount"]
    if (type(amount) ~= "number" or amount > 0)
    then
        print("ERR :" .. CName ..  " cancel --- wrong tx amount", amount)
        return false
    end
    local tb = string.split(withdrawInfo, "&")
    local receiver = tb[1]
    local assetID = tonumber(tb[2])
    local amount = tonumber(tb[3])
    if (receiver ~= sender) 
    then
        print("ERR :" .. CName ..  " cancel --- wrong sender", sender, receiver)
        return false
    end
    -- to do balance check
    L0.Transfer(receiver, assetID, amount, 0)
    L0.DelState(withdrawID)
    print("INFO:" .. CName ..  " cancel ---", withdrawID, receiver, assetID, amount)
    --]]--
    return true
end

function succeed(args)
    -- validate
    if(table.size(args) ~= 2)
    then
        print("ERR :" .. "succeed --- wrong args number", table.size(args))
        return false
    end
    -- execute
    local withdrawID = "withdraw_"..args[0]
    local feeAmount = tonumber(args[1]) 
    if (not feeAmount or feeAmount <0) 
    then
        print("ERR :" .. CName ..  " launch --- wrong feeAmount", feeAmount)
        return false
    end
    feeAmount = feeAmount * Scale
    ----[[
    local system = L0.GetState("account_system")
    local txInfo = L0.TxInfo()
    local sender = txInfo["Sender"]
    local amount = txInfo["Amount"]
    if (system ~= sender) 
    then
        print("ERR :" .. CName ..  " succeed --- wrong sender", sender, system, txInfo["Hash"])
        return false
    end
    if (type(amount) ~= "number" or amount > 0)
    then
        print("ERR :" .. CName ..  " succeed --- wrong tx amount", amount)
        return false
    end

    local withdrawInfo = L0.GetState(withdrawID)
    if (not withdrawInfo) 
    then
        print("ERR :" .. CName ..  " succeed --- withdrawID not exist", args[0])
        return false
    end
    local tb = string.split(withdrawInfo, "&")
    local assetID = tonumber(tb[2])
    local amount = tonumber(tb[3])
    if (amount < feeAmount) 
    then
        print("ERR :" .. CName ..  " succeed --- balance is not enough", feeAmount, amount)
        return false
    end
    -- to do balance check
    local fee = L0.GetState("account_fee")
    L0.Transfer(fee, assetID, feeAmount, 0)
    L0.Transfer(system, assetID, amount-feeAmount, 0)
    L0.DelState(withdrawID)
    print("INFO:" .. CName ..  " succeed ---", withdrawID, fee, assetID, feeAmount)
    print("INFO:" .. CName ..  " succeed ---", withdrawID, system, assetID, amount-feeAmount)
    --]]--
    return true
end

function fail(args)
    -- validate
    if(table.size(args) ~= 1)
    then
        print("ERR :" .. "fail --- wrong args number", table.size(args))
        return false
    end
    -- execute
    local withdrawID = "withdraw_"..args[0]
    ----[[
    local system = L0.GetState("account_system")
    local txInfo = L0.TxInfo()
    local sender = txInfo["Sender"]
    local amount = txInfo["Amount"]
    if (system ~= sender) 
    then
        print("ERR :" .. CName ..  " fail --- wrong sender", sender, system, txInfo["Hash"])
        return false
    end
    if (type(amount) ~= "number" or amount > 0)
    then
        print("ERR :" .. CName ..  " fail --- wrong tx amount", amount)
        return false
    end

    local withdrawInfo = L0.GetState(withdrawID)
    if (not withdrawInfo) 
    then
        print("ERR :" .. CName ..  " fail --- withdrawID not exist", args[0])
        return false
    end
    local tb = string.split(withdrawInfo, "&")
    local receiver = tb[1]
    local assetID = tonumber(tb[2])
    local amount = tonumber(tb[3])
    -- to do balance check
    L0.Transfer(receiver, assetID, amount,0)
    L0.DelState(withdrawID)
    --]]--
    print("INFO:" .. CName ..  " fail ---", withdrawID, receiver, assetID, amount)
    return true
end