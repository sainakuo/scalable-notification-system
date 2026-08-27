wrk.method = "POST"

wrk.body = [[
{
    "user_id": 100,
    "type": "email",
    "payload": "wrk load test"
}
]]

wrk.headers["Content-Type"] = "application/json"