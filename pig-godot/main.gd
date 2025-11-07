extends Node2D

@export var DiceOne : Sprite2D
@export var DiceTwo : Sprite2D

@export var TextLabel : RichTextLabel

func _on_roll_btn_pressed() -> void:
	DiceOne.frame_coords = Vector2i(randi()%4,randi()%3)
	DiceTwo.frame_coords = Vector2i(randi()%4,randi()%3)
	pass # Replace with function body.



func _on_pass_btn_pressed() -> void:
	pass # Replace with function body.
