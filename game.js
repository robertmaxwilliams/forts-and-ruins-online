/*
Handles the canvas and doing stuff with the server

*/
var canvas = document.getElementById('game-canvas')
var ctx = canvas.getContext('2d');
canvas.addEventListener('click', function(event) { console.log(event)}, false);


// canvas is square with a color palette on the bottom
// pal = palette because palette is hard to spell.
const boardWidth = 500
const boardHeight = 500
const palWidth = 500
const palHeight = 100

const dirtColors = ['#966341', '#9c6744', '#915d3a']

const fieldColors = ['#496645', '#5d7046', '#7d7046', '#d6b458', '#ed572d', '#9b76b8']

var gridSize = 15

function drawSquare(i, j, color){
  ctx.beginPath();
  dx = Math.floor(boardWidth / gridSize)
  dy = Math.floor(boardHeight / gridSize)
  ctx.rect(dx*i, dy*j, dx*(i+1), dy*(j+1))
  ctx.fillStyle = color
  ctx.fill();
  ctx.closePath();
}

function drawDirt(){
  let seed = 1234
  for (let i = 0; i < gridSize; i++) {
    for (let j = 0; j < gridSize; j++) {
      seed = repeatableRNG(seed)
      let color = dirtColors[seedToInt(seed, 3)]
      drawSquare(i, j, color)
    }
  }
}
drawDirt()

function repeatableRNG(seed) {
  return (seed * 9301 + 49297) % 233280
}

// there has got to be some shortcut here
// I wish I had ints
function seedToInt(seed, limit) {
 return Math.floor((limit*seed/233280) % limit)
}

function createGame(size) {
}  
