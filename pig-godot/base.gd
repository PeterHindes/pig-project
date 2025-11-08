extends Node2D

@export var game : Node2D
@export var mainMenu : Control

const address = "172.20.10.13"

func _on_join_game_pressed() -> void:
	$HTTPRequest.request_completed.connect(_on_request_completed)
	$HTTPRequest.request("http://%s:7799/game"%[address],[],HTTPClient.METHOD_POST)
func _on_request_completed(result, response_code, headers, body):
	if response_code == 200:
		var json = JSON.parse_string(body.get_string_from_utf8())
		print(json["game_id"])
		print(int(json["player_id"]))
		game.game_uuid = json["game_id"]
		game.player_index = int(json["player_id"])
		mainMenu.visible = false
		game.visible = true
		game.get_child(-1).start()
	else:
		$Control/RichTextLabel.text = "Error Connecting %d"%response_code
