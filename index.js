const loadChip8 = async () => {
  const go = new Go();
  const wasm = await fetch("chip8.wasm");
  const { instance } = await WebAssembly.instantiateStreaming(
    wasm,
    go.importObject
  );
  await go.run(instance);
};

loadChip8();

const WIDTH = 64;
const HEIGHT = 32;
const SCALE = 15;
const TICKS_PER_FRAME = 10;
let anim_frame = 0;

const canvas = document.getElementById("canvas");
canvas.width = WIDTH * SCALE;
canvas.height = HEIGHT * SCALE;

const ctx = canvas.getContext("2d");
ctx.fillStyle = "black";
ctx.fillRect(0, 0, WIDTH * SCALE, HEIGHT * SCALE);

const input = document.getElementById("fileinput");

function run() {
  document.addEventListener("keydown", function (evt) {
    EmuKeypress(evt.key, true);
  });

  document.addEventListener("keyup", function (evt) {
    EmuKeypress(evt.key, false);
  });

  input.addEventListener(
    "change",
    function (evt) {
      // Stop previous game from rendering, if one exists
      if (anim_frame != 0) {
        window.cancelAnimationFrame(anim_frame);
      }

      let file = evt.target.files[0];
      if (!file) {
        alert("Failed to read file");
        return;
      }
      // Load in game as Uint8Array, send to .wasm, start main loop
      let fr = new FileReader();
      fr.onload = function (e) {
        let buffer = fr.result;
        const rom = new Uint8Array(buffer);
        EmuReset();
        EmuLoad(rom);
        // mainloop(chip8);
      };
      fr.readAsArrayBuffer(file);
    },
    false
  );
}
run();

function mainloop() {
  // Only draw every few ticks
  for (let i = 0; i < TICKS_PER_FRAME; i++) {
    EmuTick();
  }


  // chip8.tick_timers();
  // Clear the canvas before drawing
  ctx.fillStyle = "black";
  ctx.fillRect(0, 0, WIDTH * SCALE, HEIGHT * SCALE);

  // Set the draw color back to white before we render our frame
  ctx.fillStyle = "white";

  // chip8.draw_screen(SCALE);
  EmuDrawScreen(SCALE);

  anim_frame = window.requestAnimationFrame(() => {
    mainloop();
  });
}
