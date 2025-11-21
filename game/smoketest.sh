#!/bin/bash

# Smoke test script for Generals game server
# Plays 3 games simultaneously with persistent connections
# Deps: jq, nc

set -e

SERVER_HOST="localhost"
SERVER_PORT="5000"
MAX_ACTIONS=5

log() {
    echo "[$(date '+%H:%M:%S')] $1"
}

success() {
    echo "OK $1"
}

error() {
    echo "ERR $1"
}

warning() {
    echo "WARN $1"
}

game_log() {
    local game_id="$1"
    echo "[Game $game_id] $2"
}

# Function to play a game with persistent connection
play_game() {
    local game_id="$1"
    local opponent_type="$2"
    
    game_log "$game_id" "Starting vs $opponent_type opponent"
    
    # Create temporary files for communication
    local temp_dir="/tmp/game_test_$game_id"
    mkdir -p "$temp_dir"
    local pipe_in="$temp_dir/in"
    local pipe_out="$temp_dir/out"
    
    # Create named pipes
    mkfifo "$pipe_in" "$pipe_out" 2>/dev/null || true
    
    # Start persistent nc connection in background
    nc $SERVER_HOST $SERVER_PORT < "$pipe_in" > "$pipe_out" &
    local nc_pid=$!
    
    # Give nc time to connect
    sleep 0.5
    
    # Cleanup function
    cleanup() {
        kill $nc_pid 2>/dev/null || true
        rm -rf "$temp_dir" 2>/dev/null || true
    }
    
    trap cleanup EXIT
    
    # Connect to game
    local connect_msg=$(jq -n \
        --arg type "connect" \
        --arg opponent_type "$opponent_type" \
        '{type: $type, opponent_type: $opponent_type}' -c)
    
    game_log "$game_id" "Connecting..."
    echo "$connect_msg" > "$pipe_in"
    
    # Read initial responses
    sleep 1
    local responses=""
    if [ -p "$pipe_out" ]; then
        responses=$(timeout 2 cat "$pipe_out" 2>/dev/null || echo "")
    fi
    
    # Parse initial responses
    local connected=false
    while IFS= read -r line; do
        if [[ -n "$line" ]] && echo "$line" | jq . >/dev/null 2>&1; then
            local msg_type=$(echo "$line" | jq -r '.type // "unknown"')
            
            case "$msg_type" in
                "connected")
                    local agent_id=$(echo "$line" | jq -r '.agent_id // "unknown"')
                    local opponent_id=$(echo "$line" | jq -r '.opponent_id // "unknown"')
                    local message=$(echo "$line" | jq -r '.message // ""')
                    game_log "$game_id" "Connected: $message"
                    connected=true
                    ;;
                "observation")
                    local reward=$(echo "$line" | jq -r '.reward // 0')
                    local terminated=$(echo "$line" | jq -r '.terminated // false')
                    local truncated=$(echo "$line" | jq -r '.truncated // false')
                    local timestep=$(echo "$line" | jq -r '.observation.timestep // 0')
                    local owned_land=$(echo "$line" | jq -r '.observation.owned_land_count // 0')
                    local owned_army=$(echo "$line" | jq -r '.observation.owned_army_count // 0')
                    local opponent_land=$(echo "$line" | jq -r '.observation.opponent_land_count // 0')
                    local opponent_army=$(echo "$line" | jq -r '.observation.opponent_army_count // 0')
                    
                    game_log "$game_id" "Observation: timestep=$timestep, reward=$reward"
                    game_log "$game_id" "  Your land: $owned_land, army: $owned_army"
                    game_log "$game_id" "  Opponent land: $opponent_land, army: $opponent_army"
                    
                    if [ "$terminated" = "true" ] || [ "$truncated" = "true" ]; then
                        game_log "$game_id" "Game ended!"
                        cleanup
                        return 0
                    fi
                    ;;
                "error")
                    local error_msg=$(echo "$line" | jq -r '.message // "Unknown error"')
                    game_log "$game_id" "Error: $error_msg"
                    ;;
            esac
        fi
    done <<< "$responses"
    
    if [ "$connected" != "true" ]; then
        game_log "$game_id" "Failed to connect properly"
        cleanup
        return 1
    fi
    
    # Play actions
    for action_num in $(seq 1 $MAX_ACTIONS); do
        # Generate a valid action: [to_pass, row, col, direction, to_split]
        local to_pass=0
        local row=$((RANDOM % 18))  # Grid size based on observation
        local col=$((RANDOM % 22))
        local direction=$((RANDOM % 8))  # 8 directions
        local to_split=0
        
        local action_msg=$(jq -n \
            --arg type "action" \
            --argjson action "[$to_pass, $row, $col, $direction, $to_split]" \
            '{type: $type, action: $action}' -c)
        
        game_log "$game_id" "Action $action_num/$MAX_ACTIONS: move from ($row,$col) direction $direction"
        echo "$action_msg" > "$pipe_in"
        
        # Read response
        sleep 1
        local action_response=""
        if [ -p "$pipe_out" ]; then
            action_response=$(timeout 2 cat "$pipe_out" 2>/dev/null || echo "")
        fi
        
        if [ -n "$action_response" ]; then
            while IFS= read -r line; do
                if [[ -n "$line" ]] && echo "$line" | jq . >/dev/null 2>&1; then
                    local msg_type=$(echo "$line" | jq -r '.type // "unknown"')
                    
                    case "$msg_type" in
                        "observation")
                            local reward=$(echo "$line" | jq -r '.reward // 0')
                            local terminated=$(echo "$line" | jq -r '.terminated // false')
                            local truncated=$(echo "$line" | jq -r '.truncated // false')
                            local timestep=$(echo "$line" | jq -r '.observation.timestep // 0')
                            local owned_land=$(echo "$line" | jq -r '.observation.owned_land_count // 0')
                            local owned_army=$(echo "$line" | jq -r '.observation.owned_army_count // 0')
                            local opponent_land=$(echo "$line" | jq -r '.observation.opponent_land_count // 0')
                            local opponent_army=$(echo "$line" | jq -r '.observation.opponent_army_count // 0')
                            
                            game_log "$game_id" "Observation: timestep=$timestep, reward=$reward"
                            game_log "$game_id" "  Your land: $owned_land, army: $owned_army"
                            game_log "$game_id" "  Opponent land: $opponent_land, army: $opponent_army"
                            
                            if [ "$terminated" = "true" ] || [ "$truncated" = "true" ]; then
                                game_log "$game_id" "Game ended after $action_num actions!"
                                cleanup
                                return 0
                            fi
                            ;;
                        "game_over")
                            local winner=$(echo "$line" | jq -r '.winner // "unknown"')
                            local final_rewards=$(echo "$line" | jq -r '.final_rewards // {}')
                            game_log "$game_id" "Game Over: winner=$winner, rewards=$final_rewards"
                            cleanup
                            return 0
                            ;;
                        "error")
                            local error_msg=$(echo "$line" | jq -r '.message // "Unknown error"')
                            game_log "$game_id" "Error: $error_msg"
                            ;;
                    esac
                fi
            done <<< "$action_response"
        else
            game_log "$game_id" "No response to action $action_num"
        fi
        
        # Small delay
        sleep 0.5
    done
    
    cleanup
    game_log "$game_id" "Completed $MAX_ACTIONS actions!"
}

# Main execution
main() {
    log "Starting concurrent Generals game server test"
    log "Target: $SERVER_HOST:$SERVER_PORT"
    log "Playing 3 games simultaneously, max $MAX_ACTIONS actions each"
    
    # Check dependencies
    if ! command -v jq &> /dev/null; then
        error "jq is required but not installed"
        exit 1
    fi
    
    if ! command -v nc &> /dev/null; then
        error "nc (netcat) is required but not installed"
        exit 1
    fi
    
    # Test server connectivity
    log "Testing server connectivity..."
    if ! nc -z $SERVER_HOST $SERVER_PORT 2>/dev/null; then
        error "Server is not running at $SERVER_HOST:$SERVER_PORT"
        error "Please start game server first: python main.py"
        exit 1
    fi
    success "Server is reachable"
    echo ""
    
    # Start games in background
    log "Starting 3 games in parallel..."
    
    # Game 1: vs random opponent
    play_game "1" "random" &
    local pid1=$!
    
    # Game 2: vs random opponent  
    play_game "2" "random" &
    local pid2=$!
    
    # Game 3: vs expander opponent
    play_game "3" "expander" &
    local pid3=$!
    
    # Wait for all games to complete
    local failed=0
    
    if ! wait $pid1; then
        failed=$((failed + 1))
    fi
    
    if ! wait $pid2; then
        failed=$((failed + 1))
    fi
    
    if ! wait $pid3; then
        failed=$((failed + 1))
    fi
    
    # Summary
    echo ""
    log "=== Test Summary ==="
    if [ $failed -eq 0 ]; then
        success "All games completed successfully!"
        log "Test summary: 3/3 games passed"
    else
        error "$failed games failed"
        log "Test summary: $((3 - failed))/3 games passed"
        exit 1
    fi
}

main "$@"

