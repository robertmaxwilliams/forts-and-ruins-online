/*
Handles the canvas and doing stuff with the server

*/
var mainCanvas = document.getElementById('main-canvas')
var ctx = mainCanvas.getContext('2d');

var tempCanvas = document.getElementById('temp-canvas')
var ctxTemp = tempCanvas.getContext('2d');


// canvas is square with a color palette on the bottom
// pal = palette because palette is hard to spell.
const boardWidth = 500
const boardHeight = 500
const palWidth = 500
const palHeight = 100

const colors = { 
  dirt: ['#966341', '#9c6744', '#915d3a'],
  field: ['#496645', '#5d7046', '#7d7046', '#d6b458', '#ed572d', '#9b76b8'],
  barren: '#1F2A28',
  fort: {gem: '#82355B', stone: '#A9A9A9'},
  ruin: {gem: '#691C42', stone: '#0A0A0A'},
}

var pickedColor = 0
var gridSize = 15
var dx = Math.floor(boardWidth / gridSize)
var dy = Math.floor(boardHeight / gridSize)


var boardLocal = new Array(gridSize);
for (var i = 0; i < gridSize; i++) {
  boardLocal[i] = new Array(gridSize);
}


function drawBoard(board){
  for (let i = 0; i < gridSize; i++) {
    for (let j = 0; j < gridSize; j++) {
      if (boardLocal[i][j]) {
        switch(boardLocal[i][j]) {
        case 'F': // fort
            drawSquare(i, j, colors.fort.stone, colors.fort.gem)
            break;
        case 'R': // ruin
            drawSquare(i, j, colors.ruin.stone, colors.ruin.gem)
            break;
        case 'B': // barren (dead field)
            drawSquare(i, j, colors.barren)
            break;
        default: // hopefully a number 0-6 as a string
            drawSquare(i, j, colors.field[parseInt(boardLocal[i][j])])
        }
      }
    }
  }
}

function drawSquare(i, j, color, diamondColor){

  ctx.beginPath();
  ctx.rect(dx*i, dy*j, dx, dy)
  ctx.fillStyle = color
  ctx.fill();
  ctx.closePath();
  if (diamondColor) {
    let half = Math.floor(dx/2)
    ctx.fillStyle = diamondColor
    ctx.beginPath();
    ctx.moveTo(dx*i+half, dy*j)
    ctx.lineTo(dx*i+dx, dy*j+half)
    ctx.lineTo(dx*i+half, dy*j+dy)
    ctx.lineTo(dx*i, dy*j+half)
    ctx.closePath();
    ctx.fill()
  }
}

// show preview while hovering
// only redraw when move away from square, and 
// only clear what was drawn
var lasti = -1
var lastj = -1
function hoverBoard(event){
  let x = event.offsetX
  let y = event.offsetY
  i = Math.floor(x/dx)
  j = Math.floor(y/dy)
  if (i == lasti && j == lastj) {
    return
  }
  if (boardLocal[i][j]) {
    ctxTemp.globalAlpha = 0.5
    ctxTemp.fillStyle = colors.barren
  } else {
    ctxTemp.globalAlpha = 0.8
    ctxTemp.fillStyle = colors.field[pickedColor]
  }

  ctxTemp.beginPath()
  ctxTemp.clearRect(dx*lasti, dx*lastj, dx, dy)
  ctxTemp.rect(dx*i, dx*j, dx, dy)
  ctxTemp.fill()
  let margin = 5
  ctxTemp.closePath()
  lasti = i
  lastj = j
  
 }

function clickBoard(event){

  let x = event.offsetX
  let y = event.offsetY
  console.log(x)

} 

function drawDirt(){
  let seed = 1234
  for (let i = 0; i < gridSize; i++) {
    for (let j = 0; j < gridSize; j++) {
      seed = repeatableRNG(seed)
      let color = colors.dirt[seedToInt(seed, 3)]
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

tempCanvas.addEventListener('click', clickBoard, false);
tempCanvas.addEventListener('mousemove', hoverBoard, false);
tempCanvas.addEventListener('mouseout', hoverBoard, false);


