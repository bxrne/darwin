"""
Message payload definitions for client-server communication.
All messages are JSON objects with a 'type' field.
"""

from typing import Any, Dict, Optional
from dataclasses import dataclass, asdict
import json
import numpy as np


class NumpyEncoder(json.JSONEncoder):
    """Custom JSON encoder for NumPy data types."""

    def default(self, o):
        if isinstance(o, np.ndarray):
            return o.tolist()
        if hasattr(o, "dtype") and hasattr(o, "item"):  # numpy scalar check
            if np.issubdtype(o.dtype, np.integer):
                return int(o)
            elif np.issubdtype(o.dtype, np.floating):
                return float(o)
        return super().default(o)


class MessageType:
    """Message type constants."""

    # Client -> Server
    CONNECT = "connect"
    ACTION = "action"
    RESET = "reset"
    SAVE_REPLAY = "save_replay"

    # Server -> Client
    CONNECTED = "connected"
    OBSERVATION = "observation"
    ERROR = "error"
    GAME_OVER = "game_over"


@dataclass
class ConnectRequest:
    """Client requests to join a game."""

    type: str = MessageType.CONNECT
    agent_type: str = "human"  # 'human', 'random', 'expander'
    opponent_type: str = "random"  # Type of opponent


@dataclass
class SaveReplayRequest:
    type: str = MessageType.SAVE_REPLAY


@dataclass
class ConnectedResponse:
    """Server confirms connection and provides game info."""

    type: str = MessageType.CONNECTED
    agent_id: str = ""
    opponent_id: str = ""
    message: str = "Connected to game"


@dataclass
class ActionRequest:
    """Client sends an action to perform."""

    type: str = MessageType.ACTION
    action: Any = None  # The action to perform (format depends on env)


@dataclass
class ObservationResponse:
    """Server sends observation after action."""

    type: str = MessageType.OBSERVATION
    observation: Dict[str, Any] = None
    reward: float = 0.0
    terminated: bool = False
    truncated: bool = False
    info: Dict[str, Any] = None


@dataclass
class ResetRequest:
    """Client requests game reset."""

    type: str = MessageType.RESET


@dataclass
class ErrorResponse:
    """Server sends error message."""

    type: str = MessageType.ERROR
    message: str = ""
    details: Optional[str] = None


@dataclass
class GameOverResponse:
    """Server notifies game has ended."""

    type: str = MessageType.GAME_OVER
    winner: Optional[str] = None
    final_rewards: Dict[str, float] = None
    reason: str = ""


def to_dict(obj) -> Dict[str, Any]:
    """Convert dataclass to dict, filtering None values."""
    return {k: v for k, v in asdict(obj).items() if v is not None}
