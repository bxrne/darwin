"""
Game wrapper that manages a single PettingZoo Generals game instance.
"""

import logging
from typing import Dict, Any, Optional, Tuple
from generals.agents import RandomAgent, ExpanderAgent, Agent
from generals.envs import PettingZooGenerals


class Game:
    """Manages a single game instance for a remote client."""

    def __init__(self, client_id: str, opponent_type: str = "random", render_mode: Optional[str] = None):
        """
        Initialize a game for a client.

        Args:
            client_id: Unique identifier for the client/agent
            opponent_type: Type of opponent ('random' or 'expander')
            render_mode: Optional render mode ('human', 'rgb_array', None)
        """
        self.client_id = client_id
        self.logger = logging.getLogger(f"Game-{client_id}")

        # Create opponent agent
        if opponent_type == "random":
            self.opponent = RandomAgent()
        elif opponent_type == "expander":
            self.opponent = ExpanderAgent()
        else:
            self.logger.warning(f"Unknown opponent type: {
                                opponent_type}, defaulting to random")
            self.opponent = RandomAgent()

        # Setup agent names
        self.agent_names = [client_id, self.opponent.id]

        # Initialize environment
        self.env = PettingZooGenerals(
            agents=self.agent_names,
            render_mode=render_mode
        )

        self.observations = None
        self.info = None
        self.terminated = False
        self.truncated = False

        self.logger.info(f"Game created: {client_id} vs {self.opponent.id}")

    def reset(self) -> Tuple[Dict[str, Any], Dict[str, Any]]:
        """
        Reset the game and return initial observation for client.

        Returns:
            Tuple of (observation, info) for the client agent
        """
        self.observations, self.info = self.env.reset()
        self.terminated = False
        self.truncated = False

        self.logger.info("Game reset")

        # Return client's observation
        return self.observations.get(self.client_id, {}), self.info

    def step(self, client_action: Any) -> Dict[str, Any]:
        """
        Process one game step with client's action.

        Args:
            client_action: The action from the client

        Returns:
            Dict containing observation, reward, terminated, truncated, info
        """
        if self.terminated or self.truncated:
            return self._get_terminal_state()

        try:
            # Collect actions from all agents
            actions = {}

            for agent in self.env.agents:
                if agent == self.client_id:
                    actions[agent] = client_action
                else:
                    # Opponent agent decides action
                    actions[agent] = self.opponent.act(
                        self.observations[agent])

            # Execute actions
            self.logger.info(actions)
            observations, rewards, terminated, truncated = self.env.step(
                actions)
            self.logger.info(observations)

            # Update state
            self.observations = observations
            self.terminated = terminated
            self.truncated = truncated

            # Render if enabled
            if self.env.render_mode:
                self.env.render()

            # Return client's perspective
            return {
                "observation": self.observations.get(self.client_id, {}),
                "reward": rewards.get(self.client_id, 0.0),
                "terminated": self.terminated,
                "truncated": self.truncated,
                "info": self.info
            }

        except Exception as e:
            self.logger.error(f"Error during step: {e}")
            raise

    def _get_terminal_state(self) -> Dict[str, Any]:
        """Return the terminal state for a finished game."""
        return {
            "observation": self.observations.get(self.client_id, {}) if self.observations else {},
            "reward": 0.0,
            "terminated": True,
            "truncated": self.truncated,
            "info": self.info or {}
        }

    def close(self):
        """Clean up game resources."""
        try:
            self.env.close()
            self.logger.info("Game closed")
        except Exception as e:
            self.logger.error(f"Error closing game: {e}")

    def get_winner(self) -> Optional[str]:
        """Get the winner if game is over."""
        if not (self.terminated or self.truncated):
            return None

        # Check info for winner information
        if self.info and "winner" in self.info:
            return self.info["winner"]

        return None
