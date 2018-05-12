window.onload = function () {
    var conn;
    var msg = document.getElementById("msg");
    var log = document.getElementById("log");
    var last = {};

    function appendLog(item) {
        var doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
        log.appendChild(item);
        if (doScroll) {
            log.scrollTop = log.scrollHeight - log.clientHeight;
        }
    }
    
    $("#new").click(function () {
        if (!conn) {
            return false;
        }
        
        var m = {};
        m["kind"] = "NewPost";
        m["node"] = 0;
        m["username"] = "ezfn";
        m["title"] = "test";
        m["body"] = "test post";
        m["parentHash"] = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
        m["replyToHash"] = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
        console.log(m);
        conn.send(JSON.stringify(m));
        
        return false;
    });

    $("#get").click(function () {
        if (!conn) {
            return false;
        }

        if (last === {}) {
            return false;
        }
        
        var m = {};
        m["kind"] = "GetPost";
        m["hash"] = last.Hash;
        m["parentHash"] = last.ParentHash;
        m["replyToHash"] = last.ReplyToHash;
        console.log(m);
        conn.send(JSON.stringify(m));
        
        return false;
    });
    
    if (window["WebSocket"]) {
        conn = new WebSocket("ws://" + document.location.host + "/ws");
        conn.onclose = function (evt) {
            var item = document.createElement("div");
            item.innerHTML = "<b>Connection closed.</b>";
            appendLog(item);
        };
        conn.onmessage = function (evt) {
            var m = JSON.parse(evt.data)
            msg = m.data
            console.log(m)

            if (m.Kind === "Header") {
                var item = document.createElement("div");
                item.innerText = "Header: " + msg.Username + " " + msg.Title;
                last["Hash"] = msg.Hash
                last["ParentHash"] = msg.ParentHash
                last["ReplyToHash"] = msg.ReplyToHash
                appendLog(item);
            } else if (m.Kind === "Post") {
                var item = document.createElement("div");
                item.innerText = "Post: " + msg.Username + " " + msg.Title + " " + msg.Body;
                appendLog(item);
            }
        };
    } else {
        var item = document.createElement("div");
        item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
        appendLog(item);
    }
};
