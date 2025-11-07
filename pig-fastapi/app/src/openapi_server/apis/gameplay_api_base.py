# coding: utf-8

from typing import ClassVar, Dict, List, Tuple  # noqa: F401

from pydantic import Field
from typing_extensions import Annotated
from uuid import UUID
from openapi_server.models.error_response import ErrorResponse
from openapi_server.models.game_state import GameState


class BaseGameplayApi:
    subclasses: ClassVar[Tuple] = ()

    def __init_subclass__(cls, **kwargs):
        super().__init_subclass__(**kwargs)
        BaseGameplayApi.subclasses = BaseGameplayApi.subclasses + (cls,)
    async def roll_die(
        self,
        game_id: Annotated[UUID, Field(description="The unique identifier of the game.")],
    ) -> GameState:
        ...


    async def hold_turn(
        self,
        game_id: Annotated[UUID, Field(description="The unique identifier of the game.")],
    ) -> GameState:
        ...
