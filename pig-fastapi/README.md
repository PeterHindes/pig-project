# Get setup
## getting the code to generate
```bash
# uv venv
brew install openapi-generator # macos and linux exclusive
# if you dont have brew use the pip version
# https://github.com/openAPITools/openapi-generator-pip
openapi-generator generate -i pig_game_api.yaml -g python-fastapi -o ./app
```
## getting the generated code to run
```bash
cd app
uv venv
uv pip install -r requirements.txt
source .venv/bin/activate
PYTHONPATH=src uvicorn openapi_server.main:app --host 0.0.0.0 --port 7799
```
## implementing the game logic
full path for the implementation folder
`pig-project/pig-fastapi/app/src/openapi_server/impl`
relative to the app folder
`./src/openapi_server/impl`

In this folder create a file called `game.py` or any other basic python file name.

now we need to create classes in the `game.py` file.
because of the openapi specification tags we will need two, you can see them in the pig_game_api.yaml file.
```
      tags:
        - Game Management
```
and
```
      tags:
        - Gameplay
```

these map to the classes
`class GameManagementApiImpl(BaseGameManagementApi):`
and
`class GameplayApiImpl(BaseGameplayApi):`

similarly we need to implement the functions in the classes based on the opperationId from the openapispec.
```
      operationId: roll_die
```
maps to
`async def roll_die(self, game_id: UUID) -> GameState:`
and gives us the UUID from the url of the request.
Handily the UUID validation is already done for us by FastAPI thanks to the generator.
