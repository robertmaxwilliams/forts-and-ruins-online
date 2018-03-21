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
  enemy: {gem: '#9B111E', stone: '#0A0A0A'},
  ruin: {gem: '#691C42', stone: '#0A0A0A'},
}

var pickedColor = 0
var gridSize = 15
var dx = Math.floor(boardWidth / gridSize)
var dy = Math.floor(boardHeight / gridSize)
var paldx = Math.floor(palWidth / 6)


var boardLocal = new Array(gridSize);
for (var i = 0; i < gridSize; i++) {
  boardLocal[i] = new Array(gridSize);
}

function drawBoard(board){
  for (let i = 0; i < gridSize; i++) {
    for (let j = 0; j < gridSize; j++) {
      if (boardLocal[i][j]) {
        switch(boardLocal[i][j]) {
        case username: // fort
            drawSquare(i, j, colors.fort.stone, ctx, colors.fort.gem)
            break;
        case enemyname: // enemy
            drawSquare(i, j, colors.enemy.stone, ctx, colors.enemy.gem)
            break;
        case 'R': // ruin
            drawSquare(i, j, colors.ruin.stone, ctx, colors.ruin.gem)
            break;
        case 'B': // barren (dead field)
            drawSquare(i, j, colors.barren)
            break;
        default: // hopefully a number 0-6 
            drawSquare(i, j, colors.field[parseInt(boardLocal[i][j])])
        }
      }
    }
  }
}

function drawColorPicker() {
  for (let color = 0; color < 6; color++) {
    ctx.beginPath()
    let left = Math.floor(boardWidth*(color/6))
    let width = Math.floor(boardWidth*((color+1)/6)) - left
    ctx.rect(left, boardHeight, width, palHeight)
    ctx.fillStyle = colors.field[color]
    ctx.fill()
    ctx.closePath()
  }
  drawPickedColor()
}

// draws a solid bar on top of color picker indicating what you picked
function drawPickedColor() {
  ctx.beginPath()
  ctx.rect(0, boardHeight, boardWidth, palHeight/4)
  ctx.fillStyle = colors.field[pickedColor]
  ctx.fill()
  ctx.closePath()
}
  

function drawSquare(i, j, color, context, diamondColor){

  if (!context) {
    context = ctx
  }

  let left = Math.floor(boardWidth*(i/gridSize))
  let topp = Math.floor(boardHeight*(j/gridSize))
  let width = Math.floor(boardWidth*((i+1)/gridSize)) - left
  let height = Math.floor(boardHeight*((j+1)/gridSize)) - topp

  if (color == 'erase') {
    context.clearRect(left, topp, width, height)
  } else {
    context.beginPath();
    context.rect(left, topp, width, height)
    context.fillStyle = color
    context.fill();
    context.closePath();
  }

  if (diamondColor) {
    let half = Math.floor(dx/2)
    context.fillStyle = diamondColor
    context.beginPath();
    
    let x_px = boardWidth * (i / gridSize) 
    let y_px = boardHeight * (j / gridSize)
    console.log('diamond: ', x_px, y_px)
 
    context.moveTo(x_px + half, y_px)
    context.lineTo(x_px + 2*half, y_px + half)
    context.lineTo(x_px + half, y_px + 2*half)
    context.lineTo(x_px, y_px + half)
    context.closePath();
    context.fill()
  }
}

// show preview while hovering
// only redraw when move away from square, and 
// only clear what was drawn
var lasti = -1
var lastj = -1
var lastClickedi = null
var lastClickedj = null

function hoverBoard(event){
  let x = event.offsetX
  let y = event.offsetY
  let i = Math.floor(x * gridSize / boardWidth )
  let j = Math.floor(y * gridSize / boardHeight)
  if (i == lasti && j == lastj) {
    return
  }

  // if we click on what was clicked, unclick it
  //if (i == lastClickedi && j == lastClickedj) {
   // lastClickedi = null
    //lastClickedj = null
  //}

  // if something is clicked, draw it instead
  if (lastClickedi && lastClickedj) {
    i = lastClickedi
    j = lastClickedj
  }

  // bounds checking 
  if (i < gridSize && i >= 0 && j < gridSize && j >= 0){
    if (boardLocal[i][j]) {
      ctxTemp.globalAlpha = 0.5
      ctxTemp.fillStyle = colors.barren
    } else {
      ctxTemp.globalAlpha = 1.0
      ctxTemp.fillStyle = colors.field[pickedColor]
    }

    drawSquare(i, j, colors.field[pickedColor], ctxTemp)
  }

  // don't erase if click is on
  if (!lastClickedi && !lastClickedj) {
    // clear last rect as we go
    drawSquare(lasti, lastj, 'erase', ctxTemp)
    ctxTemp.globalAlpha = 1.0
  }

  lasti = i
  lastj = j
 }

function unselect() {
  lastClickedi = null
  lastClickedj = null
}

function clickBoard(event){

  let x = event.offsetX
  let y = event.offsetY

  let i = Math.floor(x/dx)
  let j = Math.floor(y/dy)

  // if it is on the main board area
  if (i >= 0 && i < gridSize && j >= 0 && j < gridSize) {

    // if we have something clicked, unlick it
    // otherwise set down a solid click on that click
    // if it's valid, of course
    if (lastClickedi && lastClickedj) {
      console.log("unclick")
      unselect()
      socket.emit('cancelmove', null)
    } else if (!boardLocal[i][j] || boardLocal[i][j] == "") {
      console.log("click")
      lastClickedi = i 
      lastClickedj = j
      let move = {I: i, J: j, Color : pickedColor.toString()}
      socket.emit('playmove', JSON.stringify(move))
      console.log('you played', move)
    }
  // otherwise probable clicked on the pallete. or maybe an edge lol
  } else {
    pickedColor = Math.floor(x * 6 / boardWidth) % 6
    drawPickedColor()
  }

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
drawColorPicker()

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


