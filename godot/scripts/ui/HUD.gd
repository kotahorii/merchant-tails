extends CanvasLayer

@onready var day_label = $TopBar/TimeInfo/DayLabel
@onready var season_label = $TopBar/TimeInfo/SeasonLabel
@onready var weather_label = $TopBar/TimeInfo/WeatherLabel
@onready var gold_label = $TopBar/PlayerInfo/GoldLabel
@onready var reputation_label = $TopBar/PlayerInfo/ReputationLabel

func _ready():
	update_display()

func update_display():
	# This will be updated from GameMain when GDExtension provides data
	pass

func set_day(day: int):
	day_label.text = "Day: " + str(day)

func set_season(season: String):
	season_label.text = season

func set_weather(weather: String):
	weather_label.text = weather

func set_gold(amount: float):
	gold_label.text = "Gold: " + str(int(amount))

func set_reputation(rep: int):
	reputation_label.text = "Rep: " + str(rep)