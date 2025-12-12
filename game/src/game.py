"""
Game wrapper that manages a single PettingZoo Generals game instance.
"""
import numpy as np
import traceback
import logging
from typing import Dict, Any, Optional, Tuple
from generals.agents import RandomAgent, ExpanderAgent
from generals.envs import PettingZooGenerals
from generals.core.action import Action
from generals import GridFactory
from generals.core.rewards import LandRewardFn


class Game:
    """Manages a single game instance for a remote client."""

    def __init__(
        self,
        client_id: str,
        opponent_type: str = "random",
        render_mode: Optional[str] = None,
    ):
        """
        Initialize a game for a client.

        Args:
            client_id: Unique identifier for the client/agent
            opponent_type: Type of opponent ('random' or 'expander')
            render_mode: Optional render mode ('human', 'rgb_array', None)
        """
        self.client_id = client_id
        self.logger = logging.getLogger(f"Game-{client_id}")

        # Setup logger format if not already configured
        if not self.logger.handlers:
            handler = logging.StreamHandler()
            formatter = logging.Formatter(
                fmt="%(asctime)s | %(levelname)8s | %(name)s | %(message)s",
                datefmt="%Y-%m-%d %H:%M:%S",
            )
            handler.setFormatter(formatter)
            self.logger.addHandler(handler)
            self.logger.setLevel(logging.INFO)
            self.logger.propagate = False

        # Create opponent agent
        if opponent_type == "random":
            self.opponent = RandomAgent()
        elif opponent_type == "expander":
            self.opponent = ExpanderAgent()
        else:
            self.logger.warning(
                "Unknown opponent type: %s, defaulting to random", opponent_type
            )
            self.opponent = RandomAgent()

        # Setup agent names
        self.agent_names = [client_id, self.opponent.id]
        grid_factory = GridFactory(
            mode="uniform",  # Either "generalsio" or "uniform"
            # Grid height and width are randomly selected
            min_grid_dims=(8, 8),
            max_grid_dims=(8, 8),
            mountain_density=0.15,  # Probability of a mountain in a cell
            city_density=0.05,  # Probability of a city in a cell
            # Positions of generals (i, j)
            general_positions=[(2, 2), (6, 6)],
            seed=44
        )
        # Initialize environment with frequent asset rewards
        self.env = PettingZooGenerals(
            agents=self.agent_names,
            grid_factory=grid_factory,
            render_mode=render_mode,
            reward_fn=LandRewardFn(),
        )

        self.observations = None
        self.info = None
        self.terminated = False
        self.truncated = False

    def reset(self) -> Tuple[Dict[str, Any], Dict[str, Any]]:
        """
        Reset the game and return initial observation for client.

        Returns:
            Tuple of (observation, info) for the client agent
        """
        replay_name = "replays/replay" + self.client_id
        options = {"replay_file": replay_name}
        observations, info = self.env.reset(seed=44, options=options)
        self.observations = observations
        self.info = self.observations[self.client_id]["mountains"].tolist()
        self.terminated = False
        self.truncated = False

        # Return client's observation
        return extract_features(self.observations, self.client_id), self.info

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
            self.logger.debug("Received client action: %s", client_action)
            for agent in self.env.agents:
                if agent == self.client_id:
                    pass_turn = 1 if client_action[0] > 0 else 0
                    split = 1 if client_action[4] > 0 else 0

                    actions[agent] = Action(
                        pass_turn,
                        client_action[1],
                        client_action[2],
                        client_action[3],
                        split,
                    )
                else:
                    # Opponent agent decides action
                    opponent_obs = (
                        self.observations.get(
                            agent, {}) if self.observations else {}
                    )
                    actions[agent] = self.opponent.act(opponent_obs)

            # Execute actions
            observations, rewards, terminated, truncated, info = self.env.step(
                actions)
            # Update state
            self.terminated = terminated
            self.truncated = truncated
            # Render if enabled

            self.observations = observations

            # Return client's perspective
            return {
                "observation": extract_features(observations, self.client_id),
                "reward": rewards.get(self.client_id, 0.0),
                "terminated": self.terminated,
                "truncated": self.truncated,
                "info": valid_start_point_map(observations, self.client_id),
            }

        except Exception as e:
            self.logger.error("Error during step: %s", e)
            raise

    def _get_terminal_state(self) -> Dict[str, Any]:
        """Return the terminal state for a finished game."""
        return {
            "observation": self.observations.get(self.client_id, {})
            if self.observations
            else {},
            "reward": 0.0,
            "terminated": True,
            "truncated": self.truncated,
            "info": self.info or {},
        }

    def save_replay(self):
        self.env.replay.store()

    def close(self):
        """Clean up game resources."""
        try:
            self.env.close()
            self.logger.info("Game closed")
        except Exception as e:
            self.logger.error("Error closing game: %s", e)

    def get_winner(self) -> Optional[str]:
        """Get the winner if game is over."""
        if not (self.terminated or self.truncated):
            return None

        # Check info for winner information
        if self.info and "winner" in self.info:
            return self.info["winner"]

        return None


# -----------------------------------------------------------
# Helpers
# -----------------------------------------------------------


def neighbors(i, j, N, M):
    if i > 0:
        yield (i - 1, j)
    if i < N - 1:
        yield (i + 1, j)
    if j > 0:
        yield (i, j - 1)
    if j < M - 1:
        yield (i, j + 1)


def manhattan(a, b):
    return abs(a[0] - b[0]) + abs(a[1] - b[1])


# -----------------------------------------------------------
# MAIN FEATURE EXTRACTION
# state:
#  state["terrain"][i][j] → -2 mountain, -1 city, >=0 owner ID
#  state["armies"][i][j] → int army
#  state["visible"][i][j] → bool
#  state["generals"] → list of (i,j) for each player
#  state["timestep"] → int
#
# my_id = player ID you control
# -----------------------------------------------------------


def extract_features(state, my_id):
    armies = state[my_id]["armies"]
    N = len(armies)
    M = len(armies[0])
    owned_cells = state[my_id]["owned_cells"]
    opponent_cells = state[my_id]["owned_cells"]
    mountain_cells = state[my_id]["mountains"]
    cities_cells = state[my_id]["cities"]
    generals_cells = state[my_id]["generals"]
    fog_cells = state[my_id]["fog_cells"]
    owned_army = state[my_id]["owned_army_count"]
    opponent_army = state[my_id]["opponent_army_count"]
    owned_land = state[my_id]["owned_land_count"]
    opponent_land = state[my_id]["opponent_land_count"]
    fog_count = 0
    visible_cities = 0
    visible_mountains = 0
    owned_cells_list = []
    rows, cols = np.where(owned_cells)
    if rows.size == 0:
        max_army_idx = (10, 10)  # no True values
    else:
        # pick the maximum among True positions
        idx = np.argmax(armies[rows, cols])
        max_army_idx = (rows[idx], cols[idx])

    rows, cols = np.where(generals_cells & owned_cells)
    if len(rows) == 0:
        rows = [0]
        cols = [0]
    my_general_pos = (int(rows[0]), int(cols[0]))
    min_city_dist = float("inf")
    min_city_x = 0
    min_city_y = 0
    enemy_general_pos = None

    # -----------------------------
    # PASS 1: Scan the entire grid
    # -----------------------------
    for i in range(N):
        for j in range(M):
            if fog_cells[i][j]:
                fog_count += 1
                continue
            # terrain type
            if cities_cells[i][j]:  # city
                visible_cities += 1
                if not owned_cells[i][j]:
                    d = manhattan(my_general_pos, (i, j))
                    if d < min_city_dist:
                        min_city_dist = d
                        min_city_x = i
                        min_city_y = j
            elif mountain_cells[i][j]:  # mountain
                visible_mountains += 1

            # ownership
            if owned_cells[i][j]:
                owned_cells_list.append((i, j))

            elif opponent_cells[i][j]:
                if generals_cells[i][j]:
                    enemy_general_pos = (i, j)

    # -----------------------------
    # PASS 2: Border pressure
    # owned cell with at least one non-owned neighbor
    # -----------------------------
    border_pressure = 0
    for i, j in owned_cells_list:
        for ni, nj in neighbors(i, j, N, M):
            if opponent_cells[ni][nj]:
                border_pressure += 1
                break  # count each owned cell once

    # -----------------------------
    # Ratios (+1 denominator to avoid division by zero)
    # -----------------------------
    army_ratio = owned_army / (opponent_army + 1)
    land_ratio = owned_land / (opponent_land + 1)

    # -----------------------------
    # Distances (optional)
    # -----------------------------
    if enemy_general_pos is not None:
        distance_to_enemy_general = manhattan(
            my_general_pos, enemy_general_pos)
    else:
        distance_to_enemy_general = N + M  # max distance if unknown
        enemy_general_pos = (10, 10)

    # nearest visible city
    if min_city_dist == float("inf"):
        min_city_dist = N + M  # no visible city
    # -----------------------------
    # FINAL FEATURE VECTOR
    # -----------------------------
    return {
        "army_diff": owned_army - opponent_army,
        "land_diff": owned_land - opponent_land,
        "fog_count": fog_count,
        "visible_cities_count": visible_cities,
        "visible_mountains_count": visible_mountains,
        "army_ratio": float(army_ratio),
        "land_ratio": float(land_ratio),
        "border_pressure": border_pressure,
        "timestep": state[my_id]["timestep"],
        "distance_to_enemy_general": distance_to_enemy_general,
        "min_city_x": min_city_x,
        "min_city_y": min_city_y,
        "enemy_general_x": enemy_general_pos[0],
        "enemy_general_y": enemy_general_pos[1],
        "max_owned_army_x": max_army_idx[0],
        "max_owned_army_y": max_army_idx[1],

    }


def valid_start_point_map(state, my_id):
    owned_cells = state[my_id]["owned_cells"]
    army_cells = state[my_id]["armies"]

    proper_mask = owned_cells & (army_cells >= 2)
    return proper_mask.tolist()
