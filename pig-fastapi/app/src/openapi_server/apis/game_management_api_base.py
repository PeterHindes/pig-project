# coding: utf-8

from typing import ClassVar, Dict, List, Tuple  # noqa: F401

from pydantic import Field
from typing_extensions import Annotated
from uuid import UUID
from openapi_server.models.error_response import ErrorResponse
from openapi_server.models.game_state import GameState
from openapi_server.models.new_game_response import NewGameResponse


class BaseGameManagementApi:
    subclasses: ClassVar[Tuple] = ()

    def __init_subclass__(cls, **kwargs):
        super().__init_subclass__(**kwargs)
        BaseGameManagementApi.subclasses = BaseGameManagementApi.subclasses + (cls,)
    async def create_new_game(
        self,
    ) -> NewGameResponse:
        ...


    async def get_game_state(
        self,
        game_id: Annotated[UUID, Field(description="The unique identifier of the game.")],
    ) -> GameState:
        ...
