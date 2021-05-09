import 'core-js/stable';
import html from "./app.html";
import "./main.css";

const runtime = require('@wailsapp/runtime');

// Main entry point
function start() {
    // Ensure the default app div is 100% wide/high
    var app = document.getElementById('app');
    app.style.width = '100%';
    app.style.height = '100%';

    // Inject html
    app.innerHTML = html;

    let inputs = document.getElementsByClassName("inputgroup");

    for (var i = 0; i < inputs.length; i++) {
        let inputGroup = inputs[i];

        let input = inputGroup.getElementsByTagName("input")[0];
        let label = inputGroup.getElementsByTagName("label")[0];

        input.addEventListener("focus", function (e) {
            input.style.color = "#48ac62";
            label.style.color = "#48ac62";
            inputGroup.style.borderColor = "#48ac62";
        });
        input.addEventListener("blur", function (e) {
            input.style.color = "rgb(180, 180, 180)";
            label.style.color = "rgb(180, 180, 180)";
            inputGroup.style.borderColor = "rgb(180, 180, 180)";
        });
    }
    let select = document.getElementsByClassName("select");
    select[0].onclick = () => {
        backend.Controller.FilePressed()
    }

    select[1].onclick = () => {
        backend.Controller.FolderPressed()
    }

    document.getElementById("start").onclick = () => {
        startPressed();
    }

    wails.Events.On("render", data => {
        let width = data.Width;
        let height = data.Height;

        updateCanvasSize(width / height);
        var canvas = document.getElementById("render");

        let cW = canvas.width;
        let cH = canvas.height;
        var ctx = canvas.getContext("2d", { alpha: false });

        if (window.devicePixelRatio > 1) {
            var canvasWidth = canvas.width;
            var canvasHeight = canvas.height;

            canvas.width = canvasWidth * window.devicePixelRatio;
            canvas.height = canvasHeight * window.devicePixelRatio;

            canvas.style.width = canvasWidth + "px";
            canvas.style.height = canvasHeight + "px";

            ctx.scale(window.devicePixelRatio, window.devicePixelRatio);
        }

        ctx.globalCompositeOperation = "lighter";

        ctx.clearRect(0, 0, cW, cH);

        for (let tri of data.Data) {
            let c = tri.Color;
            let t = tri.Triangle.Points;
            ctx.fillStyle = `rgb(${Math.round(c.R * 255)}, ${Math.round(c.G * 255)}, ${Math.round(c.B * 255)})`;
            ctx.beginPath();
            ctx.moveTo(Math.round(t[0].X * cW), Math.round(t[0].Y * cH));
            ctx.lineTo(Math.round(t[1].X * cW), Math.round(t[1].Y * cH));
            ctx.lineTo(Math.round(t[2].X * cW), Math.round(t[2].Y * cH));
            ctx.closePath();
            ctx.fill();
        }
    });

    wails.Events.On("newPath", (path, id) => {
        addItem(path, id);
    });

    wails.Events.On("working", id => {
        document.getElementById(id).classList.add("working")
        setCheck(id)
    });

    wails.Events.On("done", id => {
        document.getElementById(id).classList.remove("working")
        document.getElementById(id).classList.add("done")
    });

    wails.Events.On("error", id => {
        let item = document.getElementById(id)
        item.classList.remove("working")
        item.classList.add("error")
        setCross(id)
    });

    wails.Events.On("time", (id, seconds) => {
        document.getElementById(id).getElementsByClassName("itemtime")[0].innerHTML = new Date(seconds * 1000).toISOString().substr(14, 5);
    });

    wails.Events.On("remove", (id) => {
        document.getElementById(id).remove();
    });
}

function updateCanvasSize(ratio) {
    var canvas = document.getElementById("render");
    var area = document.getElementById("renderContainer");

    var maxWidth = area.offsetWidth;
    var maxHeight = area.offsetHeight;
    if (maxWidth / maxHeight > ratio) {
        canvas.width = maxHeight * ratio;
        canvas.height = maxHeight;
    } else {
        canvas.width = maxWidth;
        canvas.height = maxWidth / ratio;
    }
}

function startPressed() {
    backend.Controller.StartPressed(
        parseInt(document.getElementById("points").value),
        parseInt(document.getElementById("maxtime").value),
        parseInt(document.getElementById("maxsize").value));
}

function setCross(id) {
    document.getElementById(id).getElementsByClassName("itemcontrols")[0].getElementsByTagName("button")[0].innerHTML = `
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
        <path d="M0 0h24v24H0V0z" fill="none" />
        <path
            d="M18.3 5.71c-.39-.39-1.02-.39-1.41 0L12 10.59 7.11 5.7c-.39-.39-1.02-.39-1.41 0-.39.39-.39 1.02 0 1.41L10.59 12 5.7 16.89c-.39.39-.39 1.02 0 1.41.39.39 1.02.39 1.41 0L12 13.41l4.89 4.89c.39.39 1.02.39 1.41 0 .39-.39.39-1.02 0-1.41L13.41 12l4.89-4.89c.38-.38.38-1.02 0-1.4z" />
    </svg>
    `
}

function setCheck(id) {
    document.getElementById(id).getElementsByClassName("itemcontrols")[0].getElementsByTagName("button")[0].innerHTML = `
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
        <path d="M0 0h24v24H0V0z" fill="none" />
        <path
            d="M9 16.17L5.53 12.7c-.39-.39-1.02-.39-1.41 0-.39.39-.39 1.02 0 1.41l4.18 4.18c.39.39 1.02.39 1.41 0L20.29 7.71c.39-.39.39-1.02 0-1.41-.39-.39-1.02-.39-1.41 0L9 16.17z" />
    </svg>
    `
}

function addItem(name, id) {
    if (document.getElementById("help")) {
        document.getElementById("help").remove();
    }
    var newDiv = document.createElement("div");
    newDiv.id = id;
    newDiv.classList.add("item");
    newDiv.innerHTML = `
        <div class="itemname">${name}</div>
        <div class="itemtime">00:00</div>
        <div class="itemcontrols">
            <button>
            </button>
        </div>
    `
    document.getElementById("list").appendChild(newDiv);

    document.getElementById(id).getElementsByClassName("itemcontrols")[0].getElementsByTagName("button")[0].onclick = function () {
        backend.Controller.RemoveItem(id);
    }

    setCross(id)
}

runtime.Init(start);