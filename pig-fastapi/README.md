# Get setup
## getting the code to generate
```bash
# uv venv
brew install openapi-generator # macos and linux exclusive
openapi-generator generate -i pig_game_api.yaml -g python-fastapi -o ./app
```
## getting the generated code to run
```bash
cd app
uv pip install -r requirements.txt
source .venv/bin/activate
PYTHONPATH=src uvicorn openapi_server.main:app --host 0.0.0.0 --port 7799
```
## implementing the game logic
`