//
// This code is meant as a helper to test the backend push broker.
// It is not supposed to be state-of-the-art javascript.
//

$(function() {

	var connected = false;
	var socket;

	$("#connect").click(function(){
		var service = $("#service").val();
		var A = $("#A").val();
		var M = $("#M").val();

		if(!A) {
			alert("Please choose a peer name A");
			return false;
		}

		if(connected){
			log("Closing previous socket.")
			socket.close();
		}
		socket = new WebSocket(service);
		socket.onopen = function(){
			log("Socket opened.");
       		 connected = true;

		    try {
				log("Registering peer name " + A + ".");
		        socket.send(A);
		    } catch(exception) {
		   		log(exception);
		    }
		}
		socket.onmessage = function(msg){
			log("Received message: " + msg.data);
			console.log(msg);
		}
        socket.onclose = function(){
       		 log("Socket was closed.");
       		 connected = false;
        }

	    return false;
	});

	$("#send").click(function(){
		var service = $("#service").val();
		var B = $("#B").val();
		var M = $("#M").val();

		if(!connected) {
			alert("Please connect to a server, before sending a message");
			return false;
		}
		if(!B) {
			alert("Please choose a target peer name B");
			return false;
		}
		if(!M) {
			alert("Please type a message M");
			return false;
		}

		    try {
				log("Sending message [" + M + "] to peer " + B + ".");
		        socket.send(B);
		        socket.send(M);
		    } catch(exception) {
		   		log(exception);
		    }

	    return false;
	});

	function log(msg) {
		$("#log").html( $("#log").html() + msg + "<br/>" );
	}

	var suggestions = ["Alice", "Bob", "Carol", "Eve", "Mallory", "Oscar", "Trudy", "Isaac", "David", "Susie"];

	$("#A").val( suggestions[Math.floor(Math.random() * suggestions.length)] );

});