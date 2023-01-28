var nickname, ws

function load() {
    if (!"WebSocket" in window) {
        alert("您的浏览器不支持WebSocket")
        window.opener = null
        window.open('about:blank', '_top').close()
    }
    nickname = prompt("请输入你的昵称：")
    if (nickname === null || nickname === "") {
        alert("昵称不能为空")
        window.location.href = window.location.href
    } else {
        link()
    }
}

function link() {
    ws = new WebSocket("ws://localhost:8080/ws");
    ws.onopen = function() {
        let cm = {
            "type": "setName",
            "data": nickname
        }
        ws.send(JSON.stringify(cm))
    }
    ws.onmessage = function(e) {
        let sm = JSON.parse(e.data)
        if (sm.status === 20000) {
            alert(sm.data)
            if (sm.type !== "setName") {
                return
            }
        }
        switch (sm.type) {
            case "setName":
                if (sm.status == 10000) {
                    if (banButton(document.getElementById("dislink"), 10000)) {
                        return
                    }
                    document.title = "聊天室 - " + nickname
                    document.getElementById("name").innerText = "昵称：" + nickname
                    let input = document.getElementById("input")
                    input.disabled = false
                    input.focus()
                    document.getElementById("send").hidden = false
                    document.getElementById("roll").hidden = false
                    document.getElementById("rename").hidden = false
                    document.getElementById("dislink").hidden = false
                    document.getElementById("relink").hidden = true
                } else {
                    do {
                        nickname = prompt("请输入你的昵称：")
                        if (nickname === null || nickname === "") {
                            alert("昵称不能为空")
                        } else {
                            let cm = {
                                "type": "setName",
                                "data": nickname
                            }
                            ws.send(JSON.stringify(cm))
                            break
                        }
                    } while (true);
                }
                break
            case "refreshNOP":
                document.getElementById("nop").innerText = "在线人数：" + sm.data
                break
            case "msg":
                insertOutput(sm.data)
                break
            case "send":
                input.value = ""
                break
            case "roll":
                break
            case "rename":
                nickname = sm.data
                document.title = "聊天室 - " + nickname
                document.getElementById("name").innerText = "昵称：" + nickname
                break
        }
    }
    ws.onclose = function() {
        document.title = "聊天室"
        document.getElementById("name").innerText = "未连接上服务器"
        document.getElementById("nop").innerText = ""
        document.getElementById("input").disabled = true
        document.getElementById("send").hidden = true
        document.getElementById("roll").hidden = true
        document.getElementById("rename").hidden = true
        document.getElementById("dislink").hidden = true
        document.getElementById("relink").hidden = false
        insertOutput("连接服务器失败")
    }
    ws.onerror = function(e) {
        console.log(e)
    }
}

function insertOutput(data) {
    let output = document.getElementById("output")
    if (output.value !== "") {
        output.value += "\n"
    }
    //50根短横线，通过在服务器端禁止发送多根短横线来欺骗其它人
    //原本是用的换行符，但是可以在发送的内容中使用换行符来假造信息，总不能禁掉换行符吧
    output.value += data + "\n--------------------------------------------------"
    output.scrollTop = output.scrollHeight
}

function send() {
    if (banButton(document.getElementById("send"), 1000)) {
        return
    }
    let input = document.getElementById("input")
    let cm = {
        "type": "send",
        "data": input.value
    }
    ws.send(JSON.stringify(cm))
    input.focus()
}

function roll() {
    if (banButton(document.getElementById("roll"), 10000)) {
        return
    }
    let cm = {
        "type": "roll",
        "data": ""
    }
    ws.send(JSON.stringify(cm))
}

function rename() {
    if (banButton(document.getElementById("rename"), 60000)) {
        return
    }
    let newNickname = prompt("请输入你的新昵称：")
    if (!(newNickname === null || newNickname === "")) {
        let cm = {
            "type": "rename",
            "data": newNickname
        }
        ws.send(JSON.stringify(cm))
    } else {
        document.getElementById("rename").innerText = "改名"
        document.getElementById("rename").disabled = false
    }
}

function relink() {
    if (banButton(document.getElementById("relink"), 3000)) {
        return
    }
    insertOutput("尝试重连服务器")
    link()
}

function dislink() {
    if (banButton(document.getElementById("relink"), 10000)) {
        return
    }
    insertOutput("你已断开与服务器的连接")
    ws.close()
}

function quickSend() {
    if (event.keyCode === 13 && !event.ctrlKey) {
        event.cancelBubble = true;
        event.preventDefault();
        event.stopPropagation();
        send()
    }
}

function banButton(obj, time) {
    if (obj.disabled) {
        return true
    }
    let name = obj.innerText
    obj.innerText = name + "(" + time / 1000 + "s)"
    obj.disabled = true
    setTimeout(
        function() {
            obj.innerText = name
            obj.disabled = false
        }, time)
    return false
}