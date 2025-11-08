extends Node2D

@export var game_uuid : String
@export var player_index : int
var gameOver = false
var yourTurn = false
var readyToStart = false

@export var DiceOne : Sprite2D

@export var TextLabel : RichTextLabel

const address = "172.20.10.13"

func _on_roll_btn_pressed() -> void:
	$HTTPRequest2.request_completed.disconnect(_on_request_completed_hold)
	$HTTPRequest2.request_completed.connect(_on_request_completed_roll)
	$HTTPRequest2.request("http://%s:7799/game/%s/roll"%[address,game_uuid],[],HTTPClient.METHOD_POST)
func _on_request_completed_roll(result, response_code, headers, body):
	print(response_code)
	var json = JSON.parse_string(body.get_string_from_utf8())
	print(json)
	var face = Vector2i( (int(json["last_roll"]) -1)%3 , (int(json["last_roll"]) -1)/3 )
	DiceOne.frame_coords = face

func _on_hold_btn_pressed() -> void:
	$HTTPRequest2.request_completed.disconnect(_on_request_completed_roll)
	$HTTPRequest2.request_completed.connect(_on_request_completed_hold)
	$HTTPRequest2.request("http://%s:7799/game/%s/hold"%[address,game_uuid],[],HTTPClient.METHOD_POST)
func _on_request_completed_hold(result, response_code, headers, body):
	pass

func _on_timer_timeout() -> void:
	$HTTPRequest.request_completed.connect(_on_request_completed_label)
	$HTTPRequest.request("http://%s:7799/game/%s"%[address,game_uuid],[],HTTPClient.METHOD_GET)
func _on_request_completed_label(result, response_code, headers, body):
	print(response_code)
	if response_code == 200:
		var json = JSON.parse_string(body.get_string_from_utf8())
		print(json)
		TextLabel.text = '''Game UUID: %s
%s Turn
Your Score: %d
Opponent Score: %d
Goal: %d
Other Player Joined: %s''' % [json["game_id"],'Your' if int(json["current_player_index"]) == player_index else 'Their',json["scores"][player_index],json["scores"][abs(player_index-1)],100,str(readyToStart)]
		print("Updating Label")
		readyToStart = json["ready_to_start"]
		gameOver = json["is_game_over"]
		yourTurn = (int(json["current_player_index"]) == player_index)
		# Hide buttons if its not your turn, bad security practice, but we will trust the frontend to not play for the other player when its not their turn
		$"Container/Hold Btn".visible = yourTurn and not gameOver and readyToStart
		$"Container/Roll Btn".visible = yourTurn and not gameOver and readyToStart
		
		# Display the Dice the current player is rolling
		if not typeof(json["last_roll"])==TYPE_NIL:
			var face = Vector2i( (int(json["last_roll"]) -1)%3 , (int(json["last_roll"]) -1)/3 )
			DiceOne.frame_coords = face
	$Timer.start()
