import 'core-js/stable';
const runtime = require('@wailsapp/runtime');
import html from "./app.html";

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

	document.getElementById("select").onclick = () => {
		backend.Controller.FilePressed()
	}

	wails.Events.On("newPath", (path, id) => {
		addItem(path, id);
	});
}

function addItem(name, id) {
	if (document.getElementById("help")) {
		document.getElementById("help").remove();
	}
	document.getElementById("list").innerHTML += `
    <div id=${id} class="item">
        <div class="itemname">${name}</div>
        <div class="itemtime">00:00</div>
        <div class="itemcontrols">
            <button>
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
                    <path d="M0 0h24v24H0V0z" fill="none" />
                    <path
                        d="M9 16.17L5.53 12.7c-.39-.39-1.02-.39-1.41 0-.39.39-.39 1.02 0 1.41l4.18 4.18c.39.39 1.02.39 1.41 0L20.29 7.71c.39-.39.39-1.02 0-1.41-.39-.39-1.02-.39-1.41 0L9 16.17z" />
                </svg>
            </button>
        </div>
    </div>
    `;
}

// We provide our entrypoint as a callback for runtime.Init
runtime.Init(start);