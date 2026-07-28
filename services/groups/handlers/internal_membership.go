package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"social-network/services/groups/services"
	"social-network/services/groups/utils"
)

type InternalMembershipHandlers struct {
	service *services.GroupService
}

func NewInternalMembershipHandlers(service *services.GroupService) *InternalMembershipHandlers {
	return &InternalMembershipHandlers{service: service}
}

// GetMembership handles the versioned Groups-to-service membership contract.
func (h *InternalMembershipHandlers) GetMembership(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/internal/v1/groups/"), "/")
	if len(parts) < 2 || len(parts) > 3 ||
		(parts[1] != "members" && parts[1] != "participants") ||
		(parts[1] == "participants" && len(parts) != 2) {
		utils.SendError(w, http.StatusNotFound, "Not found")
		return
	}
	groupID, err := strconv.Atoi(parts[0])
	if err != nil || groupID <= 0 {
		utils.SendError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	if parts[1] == "participants" {
		participantIDs, err := h.service.GetParticipantIDs(groupID)
		if err != nil {
			log.Printf("Failed to list group participants: %v", err)
			utils.SendError(w, http.StatusInternalServerError, "Failed to retrieve group participants")
			return
		}
		utils.SendJSON(w, http.StatusOK, map[string]any{"participant_ids": participantIDs})
		return
	}

	if len(parts) == 2 {
		memberIDs, err := h.service.GetAcceptedMemberIDs(groupID)
		if err != nil {
			log.Printf("Failed to list accepted group members: %v", err)
			utils.SendError(w, http.StatusInternalServerError, "Failed to retrieve group members")
			return
		}
		utils.SendJSON(w, http.StatusOK, map[string]any{"member_ids": memberIDs})
		return
	}

	userID, err := strconv.Atoi(parts[2])
	if err != nil || userID <= 0 {
		utils.SendError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	isMember, err := h.service.IsAcceptedMember(groupID, userID)
	if err != nil {
		log.Printf("Failed to check accepted group membership: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "Failed to check group membership")
		return
	}
	utils.SendJSON(w, http.StatusOK, map[string]any{"is_member": isMember})
}
