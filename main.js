/*
How this works:
An ajax call is made to /serverip
When it gets the IP back, actiaveSockets is called, 
then loadEveryhing is called.
You are put into a room called "lobby"

loadEveryhting checks for a cookie to log in, otherwise
prompts for log in.

You can create a new match. If someone joins the match,
you are both put into a room names yourname:theirname
Or the other was if they created the room and you joined it

Once in the room, you are both given a blank board.
Both players can make a move at the same time.
When both moves are in, the server updates the board
and sends back the changes (or maybe the whole thing, or changes+hash, git style)

When one player dies, the other wins and is happy about winning.
:wq

*/

// global vars
var username = ''
var room = 'lobby'
var otherUsers = []
var matches = []
var socket;// = io('http://localhost:8080');

// neat code to add "dif" to arrays (from https://stackoverflow.com/questions/1187518/how-to-get-the-difference-between-two-arrays-in-javascript)
// works like this: 
// [1,2].diff([2,3])
// > [1]
Array.prototype.diff = function(a) {
    return this.filter(function(i) {return a.indexOf(i) < 0;});
};

// sockets can't be loaded until server address is known, this is called by 
// ajax below.
function activateSockets(address) {
  socket = io(address);
  console.log('activated sockets')

  socket.on('connect', function() {
    console.log('connected')
  });

  socket.on('disconnection', function () {
    console.log('disocnnected')
    logout()
  });

  socket.on('userlist', function(data) {
    console.log('got user list : ', data)
    otherUsers = JSON.parse(data)
    otherUsers.sort()
    $('#user-list').text(otherUsers.join(' '))
  });

  socket.on('message', function(msg) {
    console.log(msg)
  });

  socket.on('updatedmatches', function(msg) {
    matches = JSON.parse(msg)
    console.log("got updated match list: ", msg)
    updateMatches()
  })

  socket.on('initiategame', function(gameroom) {
    console.log("someone else made gameroom: ", gameroom)
    let [host, guest] = gameroom.split(':')
    console.log("game with ", host, guest, " initiated")
    if (username == host || username == guest){
      socket.emit('room', gameroom)
      room = gameroom
      console.log('joined a game room: ', gameroom)
      //startGame(gameroom)
    }
  })
  

}

$.ajax({
  type: 'POST',
  url: 'serverip',
  success: function(data) {
    activateSockets(data)
    loadEverything()
  },
});

// jquery event driven stuff
$(document).ready(function(){
  // I wait for the server to give me a socket ip and socket to load,
  // I could wait for document ready but uh lets see what happens
})

function loadEverything(){

  username = getCookie('username')
  if (username == '') {
    logout()
  } else {
    login(username, true)
  }

  $('#login-form').submit( function( event ) {
    event.preventDefault();
    var values = {}
    $.each($(this).serializeArray(), function() {
      values[this.name] = this.value;
    });

    // login, and save cookie if remember-me is checked
    login(values['username'], (values['remember-me'] === 'on'))
    
    console.log(values)
  });

  $('#logout').click( function( event ) {
    logout(username)
    $('#login-form').show()
    $('#user-bar').hide()
  });

}

// compound functions (jquery and sockets and js)
function login(name, isSetCookie){
  socket.emit('room', room)
  console.log('joined room ', room)

  if (isSetCookie) {
    setCookie('username', name, 3)
  }
  username = name
  $('#display-username').text(username)
  $('#login-form').hide()
  $('#user-bar').show()
  $('#matchmaking').show()
  socket.emit('setname', name)
  
}


function logout(name){
  setCookie('username', '', 1)
  socket.emit('unsetname', name)
  $('#login-form').show()
  $('#user-bar').hide()
  $('#matchmaking').hide()
  username = '' 
}

function makeNewMatch(size) {
  // notice the upercase properties to agree with GoLang
  let match = {Host: username, Boardsize: size}

  // server should emit an 'updatematches' when done
  // do full reload to ensure indexs stay in sync
  console.log('making new match')
  socket.emit('makematch', JSON.stringify(match));
}

function deleteMatch(index) {
  console.log("deleting ", index)
  matches = matches.splice(index, 1)
  socket.emit('deletematch', JSON.stringify(index))
}

// creates a new room with name host:
function joinMatch(host) {
  gameroom = host + ':' + username
  // ask server to make a gameroom for us
  // it parses gameroom to get our names
  // and checks if we are in the lobby
  // then sends back 'iniategame' to both
  socket.emit('makegame', gameroom)
}
  
function updateMatches(){
  var $matchList = $("<ul>");

  matches.forEach(function (match, index) {
      let li = $("<li>")
          .text(JSON.stringify(match))
      $matchList.append(li)
      if (match.Host == username) {
        li.append( $("<button>").text("delete")
                    .on('click', function() {deleteMatch(index)}) )
      } else {
        li.append( $("<button>").text("join")
                    .on('click', function() {joinMatch(match.Host)}) )
      }
  });
  $('#match-list-container').html($matchList)
}

