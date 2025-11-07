# coding: utf-8

"""
Implementation of the Pig Game API endpoints.
This module contains the actual game logic for the Pig dice game.
"""

import random
from typing import Dict
from uuid import UUID, uuid4

from fastapi import HTTPException

from openapi_server.apis.game_management_api_base import BaseGameManagementApi
from openapi_server.apis.gameplay_api_base import BaseGameplayApi
from openapi_server.models.game_state import GameState
from openapi_server.models.new_game_response import NewGameResponse


# In-memory storage for games (in production, use a database)
game_storage: Dict[UUID, GameState] = {}

# Game configuration
WINNING_SCORE = 100


class GameManagementApiImpl(BaseGameManagementApi):
    """
    Implementation of game management operations (create, retrieve).
    """

    async def create_new_game(self) -> NewGameResponse:
        """
        Creates a new Pig game with two players.
        
        Returns:
            NewGameResponse with the newly created game_id
        """
        # Generate a new game ID
        game_id = uuid4()
        
        # Initialize a new game state
        game_state = GameState(
            game_id=game_id,
            current_player_index=0,  # Player 0 starts
            scores=[0, 0],  # Both players start with 0 points
            turn_total=0,  # No points accumulated yet
            last_roll=None,  # No roll yet
            is_game_over=False,
            winner_player_index=None
        )
        
        # Store the game
        game_storage[game_id] = game_state
        
        return NewGameResponse(game_id=game_id)

    async def get_game_state(self, game_id: UUID) -> GameState:
        """
        Retrieves the current state of a game.
        
        Args:
            game_id: The unique identifier of the game
            
        Returns:
            GameState object with current game information
            
        Raises:
            HTTPException: 404 if game not found
        """
        if game_id not in game_storage:
            raise HTTPException(status_code=404, detail=f"Game {game_id} not found")
        
        return game_storage[game_id]


class GameplayApiImpl(BaseGameplayApi):
    """
    Implementation of gameplay operations (roll, hold).
    """

    async def roll_die(self, game_id: UUID) -> GameState:
        """
        Rolls the die for the current player.
        
        Game rules:
        - Roll a die (1-6)
        - If you roll a 1: lose all turn points, turn ends automatically
        - If you roll 2-6: add to turn total, can roll again or hold
        
        Args:
            game_id: The unique identifier of the game
            
        Returns:
            Updated GameState after the roll
            
        Raises:
            HTTPException: 404 if game not found, 400 if invalid game state
        """
        # Check if game exists
        if game_id not in game_storage:
            raise HTTPException(status_code=404, detail=f"Game {game_id} not found")
        
        game = game_storage[game_id]
        
        # Check if game is already over
        if game.is_game_over:
            raise HTTPException(status_code=400, detail="Game is already over")
        
        # Roll the die (1-6)
        roll = random.randint(1, 6)
        game.last_roll = roll
        
        if roll == 1:
            # Player rolled a 1 - lose turn total and switch players
            game.turn_total = 0
            game.current_player_index = 1 - game.current_player_index
        else:
            # Add roll to turn total
            game.turn_total += roll
        
        # Update storage
        game_storage[game_id] = game
        
        return game

    async def hold_turn(self, game_id: UUID) -> GameState:
        """
        Current player holds, ending their turn.
        
        Game rules:
        - Add turn_total to player's score
        - Check for winner (score >= WINNING_SCORE)
        - Switch to next player
        - Reset turn_total to 0
        
        Args:
            game_id: The unique identifier of the game
            
        Returns:
            Updated GameState after holding
            
        Raises:
            HTTPException: 404 if game not found, 400 if invalid game state
        """
        # Check if game exists
        if game_id not in game_storage:
            raise HTTPException(status_code=404, detail=f"Game {game_id} not found")
        
        game = game_storage[game_id]
        
        # Check if game is already over
        if game.is_game_over:
            raise HTTPException(status_code=400, detail="Game is already over")
        
        # Cannot hold if last roll was a 1 (turn already ended)
        if game.last_roll == 1:
            raise HTTPException(
                status_code=400, 
                detail="Cannot hold after rolling a 1. Turn has already ended."
            )
        
        # Cannot hold if turn_total is 0 (must roll at least once)
        if game.turn_total == 0:
            raise HTTPException(
                status_code=400,
                detail="Cannot hold with zero points. You must roll at least once."
            )
        
        # Add turn total to current player's score
        current_player = game.current_player_index
        game.scores[current_player] += game.turn_total
        
        # Check if current player won
        if game.scores[current_player] >= WINNING_SCORE:
            game.is_game_over = True
            game.winner_player_index = current_player
            game.turn_total = 0
            game.last_roll = None
        else:
            # Switch to next player
            game.current_player_index = 1 - game.current_player_index
            game.turn_total = 0
            game.last_roll = None
        
        # Update storage
        game_storage[game_id] = game
        
        return game
