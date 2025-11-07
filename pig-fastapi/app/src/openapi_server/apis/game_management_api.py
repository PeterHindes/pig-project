# coding: utf-8

from typing import Dict, List  # noqa: F401
import importlib
import pkgutil

from openapi_server.apis.game_management_api_base import BaseGameManagementApi
import openapi_server.impl

from fastapi import (  # noqa: F401
    APIRouter,
    Body,
    Cookie,
    Depends,
    Form,
    Header,
    HTTPException,
    Path,
    Query,
    Response,
    Security,
    status,
)

from openapi_server.models.extra_models import TokenModel  # noqa: F401
from pydantic import Field
from typing_extensions import Annotated
from uuid import UUID
from openapi_server.models.error_response import ErrorResponse
from openapi_server.models.game_state import GameState
from openapi_server.models.new_game_response import NewGameResponse


router = APIRouter()

ns_pkg = openapi_server.impl
for _, name, _ in pkgutil.iter_modules(ns_pkg.__path__, ns_pkg.__name__ + "."):
    importlib.import_module(name)


@router.post(
    "/game",
    responses={
        201: {"model": NewGameResponse, "description": "Game created successfully."},
        500: {"model": ErrorResponse, "description": "Internal server error."},
    },
    tags=["Game Management"],
    summary="Start a new Pig game.",
    response_model_by_alias=True,
)
async def create_new_game(
) -> NewGameResponse:
    if not BaseGameManagementApi.subclasses:
        raise HTTPException(status_code=500, detail="Not implemented")
    return await BaseGameManagementApi.subclasses[0]().create_new_game()


@router.get(
    "/game/{game_id}",
    responses={
        200: {"model": GameState, "description": "Current game state."},
        404: {"model": ErrorResponse, "description": "Game not found."},
        500: {"model": ErrorResponse, "description": "Internal server error."},
    },
    tags=["Game Management"],
    summary="Get the current state of a specific game.",
    response_model_by_alias=True,
)
async def get_game_state(
    game_id: Annotated[UUID, Field(description="The unique identifier of the game.")] = Path(..., description="The unique identifier of the game."),
) -> GameState:
    if not BaseGameManagementApi.subclasses:
        raise HTTPException(status_code=500, detail="Not implemented")
    return await BaseGameManagementApi.subclasses[0]().get_game_state(game_id)
