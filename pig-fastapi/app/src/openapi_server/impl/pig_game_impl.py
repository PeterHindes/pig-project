# coding: utf-8

"""
Implementation of the Pig Game API endpoints.
This module contains the actual game logic for the Pig dice game with player matchmaking.
"""

import random
from typing import Dict
from uuid import UUID, uuid4

from fastapi import HTTPException
from pydantic import BaseModel

from openapi_server.apis.game_management_api_base import BaseGameManagementApi
from openapi_server.apis.gameplay_api_base import BaseGameplayApi
from openapi_server.models.game_state import GameState
from openapi_server.models.new_game_response import NewGameResponse


# Game metadata to track player count
class GameMetadata(BaseModel):
    """Metadata about a game including its state and player count"""
    state: GameState
    player_count: int  # Number of players who have joined (1 or 2)


# In-memory storage for games (in production, use a database)
game_storage: Dict[UUID, GameMetadata] = {}

# Game configuration
WINNING_SCORE = 100


class GameManagementApiImpl(BaseGameManagementApi):
    """
    Implementation of game management operations with player matchmaking.
    """

    async def create_new_game(self) -> NewGameResponse:
        """
        Creates a new Pig game or joins an existing game waiting for a second player.
        
        Matchmaking logic:
        1. Look for any existing game with only 1 player
        2. If found, add this caller as player 1 (index 1)
        3. If not found, create a new game with this caller as player 0 (index 0)
        
        Returns:
            NewGameResponse with game_id and player_id (0 or 1)
        """
        # Look for a game waiting for a second player
        waiting_game_id = None
        for game_id, metadata in game_storage.items():
            if metadata.player_count == 1 and not metadata.state.is_game_over:
                waiting_game_id = game_id
                break
        
        if waiting_game_id:
            # Join existing game as player 1
            game_metadata = game_storage[waiting_game_id]
            old_state = game_metadata.state
            
            # Create new state with ready_to_start = True
            new_state = GameState(
                game_id=old_state.game_id,
                current_player_index=old_state.current_player_index,
                scores=old_state.scores,
                turn_total=old_state.turn_total,
                last_roll=old_state.last_roll,
                ready_to_start=True,  # Now we have 2 players
                is_game_over=old_state.is_game_over,
                winner_player_index=old_state.winner_player_index
            )
            
            game_metadata.state = new_state
            game_metadata.player_count = 2
            game_storage[waiting_game_id] = game_metadata
            
            # Return with player_id = 1
            return NewGameResponse(
                game_id=waiting_game_id,
                player_id=1
            )
        
        else:
            # Create a new game with this caller as player 0
            game_id = uuid4()
            
            # Initialize a new game state
            game_state = GameState(
                game_id=game_id,
                current_player_index=0,  # Player 0 starts
                scores=[0, 0],  # Both players start with 0 points
                turn_total=0,  # No points accumulated yet
                last_roll=None,  # No roll yet
                ready_to_start=False,  # Waiting for player 1 to join
                is_game_over=False,
                winner_player_index=None
            )
            
            # Store the game with metadata
            game_metadata = GameMetadata(
                state=game_state,
                player_count=1  # Only player 0 has joined so far
            )
            game_storage[game_id] = game_metadata
            
            # Return with player_id = 0
            return NewGameResponse(
                game_id=game_id,
                player_id=0
            )

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
        
        return game_storage[game_id].state


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
        
        game_metadata = game_storage[game_id]
        old_state = game_metadata.state
        
        # Check if game is ready to start
        if not old_state.ready_to_start:
            raise HTTPException(
                status_code=400, 
                detail="Cannot start game. Waiting for second player to join."
            )
        
        # Check if game is already over
        if old_state.is_game_over:
            raise HTTPException(status_code=400, detail="Game is already over")
        
        # Roll the die (1-6)
        roll = random.randint(1, 6)
        
        # Create new state with the roll result
        if roll == 1:
            # Player rolled a 1 - lose turn total and switch players
            new_state = GameState(
                game_id=old_state.game_id,
                current_player_index=1 - old_state.current_player_index,
                scores=old_state.scores.copy(),
                turn_total=0,
                last_roll=roll,  # Explicitly set the roll value
                ready_to_start=old_state.ready_to_start,
                is_game_over=old_state.is_game_over,
                winner_player_index=old_state.winner_player_index
            )
        else:
            # Add roll to turn total
            new_state = GameState(
                game_id=old_state.game_id,
                current_player_index=old_state.current_player_index,
                scores=old_state.scores.copy(),
                turn_total=old_state.turn_total + roll,
                last_roll=roll,  # Explicitly set the roll value
                ready_to_start=old_state.ready_to_start,
                is_game_over=old_state.is_game_over,
                winner_player_index=old_state.winner_player_index
            )
        
        # Update storage with new state
        game_metadata.state = new_state
        game_storage[game_id] = game_metadata
        
        return new_state

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
        
        game_metadata = game_storage[game_id]
        old_state = game_metadata.state
        
        # Check if game is ready to start
        if not old_state.ready_to_start:
            raise HTTPException(
                status_code=400,
                detail="Cannot start game. Waiting for second player to join."
            )
        
        # Check if game is already over
        if old_state.is_game_over:
            raise HTTPException(status_code=400, detail="Game is already over")
        
        # Cannot hold if last roll was a 1 (turn already ended)
        if old_state.last_roll == 1:
            raise HTTPException(
                status_code=400, 
                detail="Cannot hold after rolling a 1. Turn has already ended."
            )
        
        # Cannot hold if turn_total is 0 (must roll at least once)
        if old_state.turn_total == 0:
            raise HTTPException(
                status_code=400,
                detail="Cannot hold with zero points. You must roll at least once."
            )
        
        # Add turn total to current player's score
        current_player = old_state.current_player_index
        new_scores = old_state.scores.copy()
        new_scores[current_player] += old_state.turn_total
        
        # Check if current player won
        if new_scores[current_player] >= WINNING_SCORE:
            new_state = GameState(
                game_id=old_state.game_id,
                current_player_index=current_player,
                scores=new_scores,
                turn_total=0,
                last_roll=None,  # Reset for game over
                ready_to_start=old_state.ready_to_start,
                is_game_over=True,
                winner_player_index=current_player
            )
        else:
            # Switch to next player
            new_state = GameState(
                game_id=old_state.game_id,
                current_player_index=1 - current_player,
                scores=new_scores,
                turn_total=0,
                last_roll=None,  # Reset for next turn
                ready_to_start=old_state.ready_to_start,
                is_game_over=old_state.is_game_over,
                winner_player_index=old_state.winner_player_index
            )
        
        # Update storage with new state (FIXED: store metadata, not just state)
        game_metadata.state = new_state
        game_storage[game_id] = game_metadata
        
        return new_state
