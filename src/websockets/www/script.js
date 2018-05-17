var conn;
var node = -1;
var requestedHash;

var blankHash = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";

function _atob(buffer) {
    var binary = '';
    var bytes = new Uint8Array( buffer );
    var len = bytes.byteLength;
    for (var i = 0; i < len; i++) {
        binary += String.fromCharCode( bytes[ i ] );
    }
    return window.btoa( binary );
}

function showPost(title, username, hash, parentHash, replyToHash) {
    var item = `<div id="${hash}-p" class="post"><strong>${title}</strong> (posted by ${username})&nbsp;&nbsp;
    <button type="button" id="${hash}-get" class="btn btn-outline-primary btn-sm"><i class="fa fa-download"></i></button>
    <button type="button" id="${hash}-reply" class="btn btn-outline-primary btn-sm" style="display: none;"><i class="fa fa-reply"></i></button><br>
    <span id="${hash}-body"></span>
    <div class="comment-reply" id="${hash}-comment" style="display: none;">
      <hr>
      <div class="form-inline" style="margin-bottom: 10px;">
        <input type="text" class="form-control" id="${hash}-commentUsername" placeholder="Username">
        <input type="text" class="form-control" id="${hash}-commentInput" placeholder="Reply">
      </div>
      <button type="button" class="btn btn-secondary btn-sm" id="${hash}-commentClose">Close</button>
      <button type="button" class="btn btn-primary btn-sm" id="${hash}-commentSubmit">Submit comment</button>
    </div>
    </div>`;

    if (replyToHash != blankHash) {
        // It's a comment.
        $(`[id="${replyToHash}-p"]`).append(item);
    } else {
        // It's a fresh post, not a comment.
        $("#container").prepend(item);
    }

    $(`[id="${hash}-get"]`).click(function () {
        getPost(hash, parentHash, replyToHash);
    });

    $(`[id="${hash}-reply"]`).click(function () {
        $(`[id="${hash}-comment"]`).show();
    });

    $(`[id="${hash}-commentClose"]`).click(function () {
        $(`[id="${hash}-comment"]`).hide();
    });

    $(`[id="${hash}-commentSubmit"]`).click(function () {
        sendPost("reply", $(`[id="${hash}-commentUsername"]`).val(), $(`[id="${hash}-commentInput"]`).val(), "", hash)
        $(`[id="${hash}-comment"]`).hide();
    });
}

function fillPost(body, hash) {
    $(`[id="${hash}-body"]`).html(`${body}<br>`);
    $(`[id="${hash}-get"]`).hide();
    $(`[id="${hash}-reply"]`).show();
}

function sendPost(title, username, body, parentHash, replyToHash) {
    if (!conn || conn.readyState != 1) {
        return;
    }

    if (title === "" || username === "" || body === "") {
        return;
    }

    if (parentHash === "") {
        parentHash = blankHash; 
    }

    if (replyToHash === "") {
        replyToHash = blankHash; 
    }

    var m = {
        kind: "NewPost",
        node: node,
        title: title,
        username: username,
        body: body,
        parentHash: parentHash,
        replyToHash: replyToHash
    };

    console.log(m);
    conn.send(JSON.stringify(m));
}

function getPost(hash, parentHash, replyToHash) {
    if (!conn || conn.readyState != 1) {
        return;
    }

    var m = {
        kind: "GetPost",
        node: node,
        hash: hash,
        parentHash: parentHash,
        replyToHash: replyToHash
    };

    requestedHash = hash;
    console.log(m);
    conn.send(JSON.stringify(m)); 
}

$(document).ready(function () {
    if (window["WebSocket"]) {
        conn = new WebSocket("ws://localhost:8080/ws");
        conn.onclose = function (evt) {
            showPost("Connection closed; try refreshing.", "null", "", "", blankHash);
        };
        conn.onmessage = function (evt) {
            var m = JSON.parse(evt.data);
            msg = m.Data;
            console.log(m);
            if (m.Kind === "Header") {
                showPost(msg.Title, msg.Username, _atob(msg.Seed.Hash), _atob(msg.Seed.ParentHash), _atob(msg.Seed.ReplyToHash));
            } else if (m.Kind === "Post") {
                fillPost(msg.Body, requestedHash);
            } else if (m.Kind === "Setup") {
                node = m.Node
                $("#node").text(node.toString());
            }
        };
    } else {
            showPost("Your browser does not support WebSockets.", "null", "", "", blankHash);
    }

    setTimeout(function() {
        console.log('test2', node)
        if (node == -1) {
            showPost("We're out of nodes :( Try refreshing in a bit.", "null", "", "", blankHash);
        }
    }, 2000);

    $("#submitNewPost").click(function () {
        console.log("test");
        sendPost($("#titleInput").val(), $("#usernameInput").val(), $("#bodyInput").val(), "", "");
        $("#titleInput").val("");
        $("#usernameInput").val("");
        $("#bodyInput").val("");
        $('#newPostModal').modal('toggle');
    });
});