package notification

import (
	"context"
	"strconv"

	"github.com/DANazavr/RATest/internal/common/meta"
	"github.com/DANazavr/RATest/internal/domain/models"
	"github.com/DANazavr/RATest/internal/log"
	"github.com/DANazavr/RATest/internal/services"
	"github.com/DANazavr/RATest/protos/gen/go/notification"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NotificationServer struct {
	ctx                 context.Context
	logger              *log.Log
	userService         *services.UserService
	notificationService *services.NotificationService
	notification.UnimplementedNotificationServer
}

func NewNotificationServer(ctx context.Context, logger *log.Log, us *services.UserService, ns *services.NotificationService) *NotificationServer {
	return &NotificationServer{
		ctx:                 ctx,
		logger:              logger.WithComponent("grpc/notification/notificationServer"),
		userService:         us,
		notificationService: ns,
	}
}

func Register(gRPC *grpc.Server, notificationServer *NotificationServer) {
	notification.RegisterNotificationServer(gRPC, notificationServer)
}

func (ns *NotificationServer) Publish(ctx context.Context, req *notification.PublishRequest) (*notification.PublishResponse, error) {
	var userID int
	for i, v := range req.Channel {
		if v == '#' {
			var err error
			userID, err = strconv.Atoi(req.Channel[i+1:])
			if err != nil {
				ns.logger.Errorf(ns.ctx, "Failed to decode userid: %v", err)
			}
		}
	}

	if userID <= 0 {
		ns.logger.Errorf(ns.ctx, "Invalid user ID: %d", userID)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid user ID provided in channel")
	}

	if _, err := ns.userService.UsersGetById(userID); err != nil {
		ns.logger.Errorf(ns.ctx, "Failed to get user: %v", err)
		return nil, status.Errorf(codes.NotFound, "User with ID %d not found", userID)
	}

	notificationMap := map[string]interface{}{
		"title":   req.Data.Title,
		"message": req.Data.Message,
	}
	n := &models.UserNotification{
		UserID:       userID,
		Notification: notificationMap,
	}

	if err := ns.notificationService.NotificationCreate(n); err != nil {
		ns.logger.Errorf(ns.ctx, "Failed to create notification: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to create notification: %v", err)
	}
	ns.logger.Infof(ns.ctx, "Notification created for user %d: %v", userID, n.UID)

	presence, err := ns.notificationService.Presence(req.Channel)
	if err != nil {
		ns.logger.Errorf(ns.ctx, "Failed to get presence: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to get presence: %v", err)
	}
	if len(presence.Presence) == 0 {
		ns.logger.Infof(ns.ctx, "No users online in channel %s", req.Channel)
	}

	publish, err := ns.notificationService.Publish(n, req.Channel)
	if err != nil {
		ns.logger.Errorf(ns.ctx, "Failed to publish notification: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to publish notification: %v", err)
	}

	for _, v := range presence.Presence {
		userid, err := strconv.Atoi(v.User)
		if err != nil {
			ns.logger.Errorf(ns.ctx, "Failed to convert user ID from presence: %v", err)
			return nil, status.Errorf(codes.Internal, "Failed to convert user ID from presence: %v", err)
		}
		if err := ns.notificationService.MarkAsSend(n, userid); err != nil {
			ns.logger.Errorf(ns.ctx, "Failed to mark notification as sent: %v", err)
			return nil, status.Errorf(codes.Internal, "Failed to mark notification as sent: %v", err)
		}
	}
	return &notification.PublishResponse{Offset: publish.Offset, Epoch: publish.Epoch}, nil
}

func (ns *NotificationServer) Broadcast(ctx context.Context, req *notification.BroadcastRequest) (*notification.BroadcastResponse, error) {
	users, err := ns.userService.UsersGet()
	if err != nil {
		ns.logger.Errorf(ns.ctx, "Failed to get users: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to get users: %v", err)
	}
	for _, user := range users {
		if user.Role != "user" {
			ns.logger.Infof(ns.ctx, "Skipping user %d with role %s", user.ID, user.Role)
			continue
		}
		notificationMap := map[string]interface{}{
			"title":   req.Data.Title,
			"message": req.Data.Message,
		}
		n := &models.UserNotification{
			UserID:       user.ID,
			Notification: notificationMap,
		}

		channel := "notifications:user#" + strconv.Itoa(user.ID)

		if err := ns.notificationService.NotificationCreate(n); err != nil {
			ns.logger.Errorf(ns.ctx, "Failed to create notification: %v", err)
			return nil, status.Errorf(codes.Internal, "Failed to create notification: %v", err)
		}
		ns.logger.Infof(ns.ctx, "Notification created for user %d: %v", user.ID, n.UID)
		_, err := ns.notificationService.Publish(n, channel)
		if err != nil {
			ns.logger.Errorf(ns.ctx, "Failed to publish notification: %v", err)
			return nil, status.Errorf(codes.Internal, "Failed to publish notification: %v", err)
		}
		presence, err := ns.notificationService.Presence(channel)
		if err != nil {
			ns.logger.Errorf(ns.ctx, "Failed to get presence: %v", err)
			return nil, status.Errorf(codes.Internal, "Failed to get presence: %v", err)
		}
		if len(presence.Presence) == 0 {
			ns.logger.Infof(ns.ctx, "No users online in channel %s", channel)
			return nil, status.Errorf(codes.NotFound, "No users online in channel %s, but message is publish", channel)
		}
		for _, v := range presence.Presence {
			userid, err := strconv.Atoi(v.User)
			if err != nil {
				ns.logger.Errorf(ns.ctx, "Failed to convert user ID from presence: %v", err)
				return nil, status.Errorf(codes.Internal, "Failed to convert user ID from presence: %v", err)
			}
			if err := ns.notificationService.MarkAsSend(n, userid); err != nil {
				ns.logger.Errorf(ns.ctx, "Failed to mark notification as sent: %v", err)
				return nil, status.Errorf(codes.Internal, "Failed to mark notification as sent: %v", err)
			}
		}
	}
	ns.logger.Infof(ctx, "Broadcast notification sent: %v", req.Data)
	return &notification.BroadcastResponse{Message: "Broadcast notification sent successfully"}, nil
}

func (ns *NotificationServer) MarkAsRead(ctx context.Context, req *notification.MarkAsReadRequest) (*notification.MarkAsReadResponse, error) {
	userIDstr, ok := ctx.Value(meta.UserIDKey).(string)
	if !ok || userIDstr == "" {
		ns.logger.Errorf(ctx, "Invalid user ID in context: %v", userIDstr)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid user ID in context: %v", userIDstr)
	}
	userID, err := strconv.Atoi(userIDstr)
	if err != nil {
		ns.logger.Errorf(ns.ctx, "Failed to convert user ID to int: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid user ID: %v", userIDstr)
	}
	notificationID, err := strconv.Atoi(req.NotificationId)
	if err != nil {
		ns.logger.Errorf(ns.ctx, "Failed to convert user ID to int: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid user ID: %v", userIDstr)
	}

	_, err = ns.userService.UsersGetById(userID)
	if err != nil {
		ns.logger.Errorf(ns.ctx, "Failed to get user: %v", err)
		return nil, status.Errorf(codes.NotFound, "User with ID %d not found", userID)
	}
	n, err := ns.notificationService.GetById(notificationID)
	if err != nil {
		ns.logger.Errorf(ns.ctx, "Failed to get notifications: %v", err)
		return nil, status.Errorf(codes.NotFound, "Notification with ID %d not found", notificationID)
	}

	if err := ns.notificationService.MarkAsRead(n, userID); err != nil {
		ns.logger.Errorf(ns.ctx, "Failed to mark notification as read: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to mark notification as read: %v", err)
	}
	return &notification.MarkAsReadResponse{Message: "Notification marked as read successfully"}, nil
}

func (ns *NotificationServer) GetNotificationsByFilter(ctx context.Context, req *notification.GetNotificationsByFilterRequest) (*notification.GetNotificationsByFilterResponse, error) {
	userIDstr, ok := ctx.Value(meta.UserIDKey).(string)
	if !ok || userIDstr == "" {
		ns.logger.Errorf(ns.ctx, "Invalid user ID in context: %v", userIDstr)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid user ID in context: %v", userIDstr)
	}
	userID, err := strconv.Atoi(userIDstr)
	if err != nil {
		ns.logger.Errorf(ns.ctx, "Failed to convert user ID to int: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid user ID: %v", userIDstr)
	}

	_, err = ns.userService.UsersGetById(userID)
	if err != nil {
		ns.logger.Errorf(ns.ctx, "Failed to get user: %v", err)
		return nil, status.Errorf(codes.NotFound, "User with ID %d not found", userID)
	}

	filter := req.Filter
	if !ns.notificationService.ValidateFilter(filter) {
		ns.logger.Errorf(ns.ctx, "Invalid filter: %s", filter)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid filter: %s", filter)
	}

	notifications, err := ns.notificationService.GetByUserIdWithFilter(userID, filter)
	if err != nil {
		ns.logger.Errorf(ns.ctx, "Failed to get notifications: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to get notifications: %v", err)
	}

	protoNotifications := make([]*notification.Notification, 0, len(notifications))
	for _, v := range notifications {
		if v.SendAt == nil {
			if err := ns.notificationService.MarkAsSend(v, userID); err != nil {
				ns.logger.Errorf(ns.ctx, "Failed to mark notification as sent: %v", err)
				return nil, status.Errorf(codes.Internal, "Failed to mark notification as sent: %v", err)
			}
		}
		protoNotif, err := ns.notificationService.ConvertToProtoNotification(v)
		if err != nil {
			ns.logger.Errorf(ns.ctx, "Failed to convert notification: %v", err)
			return nil, status.Errorf(codes.Internal, "Failed to convert notification: %v", err)
		}
		protoNotifications = append(protoNotifications, protoNotif)
	}

	return &notification.GetNotificationsByFilterResponse{Notifications: protoNotifications}, nil
}
