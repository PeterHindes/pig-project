# coding: utf-8

from fastapi.testclient import TestClient


from pydantic import Field  # noqa: F401
from typing_extensions import Annotated  # noqa: F401
from uuid import UUID  # noqa: F401
from openapi_server.models.error_response import ErrorResponse  # noqa: F401
from openapi_server.models.game_state import GameState  # noqa: F401
from openapi_server.models.new_game_response import NewGameResponse  # noqa: F401


def test_create_new_game(client: TestClient):
    """Test case for create_new_game

    Start a new Pig game.
    """

    headers = {
    }
    # uncomment below to make a request
    #response = client.request(
    #    "POST",
    #    "/game",
    #    headers=headers,
    #)

    # uncomment below to assert the status code of the HTTP response
    #assert response.status_code == 200


def test_get_game_state(client: TestClient):
    """Test case for get_game_state

    Get the current state of a specific game.
    """

    headers = {
    }
    # uncomment below to make a request
    #response = client.request(
    #    "GET",
    #    "/game/{game_id}".format(game_id='game_id_example'),
    #    headers=headers,
    #)

    # uncomment below to assert the status code of the HTTP response
    #assert response.status_code == 200

