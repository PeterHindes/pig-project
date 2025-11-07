# coding: utf-8

from typing import Dict, List  # noqa: F401
import importlib
import pkgutil

from openapi_server.apis.gameplay_api_base import BaseGameplayApi
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


router = APIRouter()

ns_pkg = openapi_server.impl
for _, name, _ in pkgutil.iter_modules(ns_pkg.__path__, ns_pkg.__name__ + "."):
    importlib.import_module(name)


@router.post(
    "/game/{game_id}/roll",
    responses={
        200: {"model": GameState, "description": "Game state after rolling the die."},
        400: {"model": ErrorResponse, "description": "Invalid game state (e.g., game already over)."},
        404: {"model": ErrorResponse, "description": "Game not found."},
        500: {"model": ErrorResponse, "description": "Internal server error."},
    },
    tags=["Gameplay"],
    summary="Roll the die for the current player.",
    response_model_by_alias=True,
)
async def roll_die(
    game_id: Annotated[UUID, Field(description="The unique identifier of the game.")] = Path(..., description="The unique identifier of the game."),
) -> GameState:
    if not BaseGameplayApi.subclasses:
        raise HTTPException(status_code=500, detail="Not implemented")
    return await BaseGameplayApi.subclasses[0]().roll_die(game_id)


@router.post(
    "/game/{game_id}/hold",
    responses={
        200: {"model": GameState, "description": "Game state after holding."},
        400: {"model": ErrorResponse, "description": "Invalid game state (e.g., game already over, cannot hold after rolling a 1)."},
        404: {"model": ErrorResponse, "description": "Game not found."},
        500: {"model": ErrorResponse, "description": "Internal server error."},
    },
    tags=["Gameplay"],
    summary="Current player holds, ending their turn and adding turn total to score.",
    response_model_by_alias=True,
)
async def hold_turn(
    game_id: Annotated[UUID, Field(description="The unique identifier of the game.")] = Path(..., description="The unique identifier of the game."),
) -> GameState:
    if not BaseGameplayApi.subclasses:
        raise HTTPException(status_code=500, detail="Not implemented")
    return await BaseGameplayApi.subclasses[0]().hold_turn(game_id)
