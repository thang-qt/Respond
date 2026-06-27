package i18n

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	DefaultLocale = "en"
	LocaleEN      = "en"
	LocaleVI      = "vi"
)

func NormalizeLocale(locale string) string {
	locale = strings.TrimSpace(strings.ToLower(locale))
	locale = strings.ReplaceAll(locale, "_", "-")
	if strings.HasPrefix(locale, "vi") {
		return LocaleVI
	}
	if strings.HasPrefix(locale, "en") {
		return LocaleEN
	}
	return DefaultLocale
}

func LocaleFromRequest(r *http.Request) string {
	if r == nil {
		return DefaultLocale
	}
	if cookie, err := r.Cookie("NEXT_LOCALE"); err == nil && cookie != nil && cookie.Value != "" {
		return NormalizeLocale(cookie.Value)
	}
	return NormalizeLocale(r.Header.Get("Accept-Language"))
}

type Vars map[string]any

func T(locale, key string, vars Vars) string {
	catalog := messages[NormalizeLocale(locale)]
	if catalog == nil {
		catalog = messages[DefaultLocale]
	}
	template, ok := catalog[key]
	if !ok {
		template = messages[DefaultLocale][key]
	}
	if template == "" {
		return key
	}
	for name, value := range vars {
		template = strings.ReplaceAll(template, "{"+name+"}", fmt.Sprint(value))
	}
	return template
}

var messages = map[string]map[string]string{
	LocaleEN: {
		"error.server":              "Something went wrong.",
		"error.authRequired":        "Authentication required.",
		"error.mustSignIn":          "You must be signed in.",
		"error.verifyEmail":         "Verify your email to perform this action.",
		"error.permissionDenied":    "You do not have permission to perform this action.",
		"error.accountSuspended":    "This account is temporarily suspended.",
		"error.accountBanned":       "This account is permanently banned.",
		"error.invalidRequestBody":  "Invalid request body.",
		"error.invalidPage":         "Invalid page.",
		"error.invalidPerPage":      "Invalid per_page.",
		"error.validation":          "Invalid request.",
		"error.debateNotFound":      "This debate doesn't exist or has been removed.",
		"error.debateHiddenByBlock": "This debate is hidden by your safety settings.",
		"error.localeInvalid":       "locale must be one of: en, vi.",

		"notification.challenge.expired":          "Your challenge for \"{topic}\" expired.",
		"notification.challenge.declined":         "Your challenge for \"{topic}\" was declined.",
		"notification.challenge.accepted":         "Your challenge for \"{topic}\" was accepted.",
		"notification.challenge.invited":          "@{username} invited you to join: \"{topic}\"",
		"notification.challenge.challenged":       "@{username} challenged you: \"{topic}\"",
		"notification.challenge.previousOpponent": "Your previous opponent challenged you: \"{topic}\"",
		"notification.comment.reply":              "Someone replied to your reflection in \"{topic}\"",
		"notification.debate.joined":              "Someone joined your debate \"{topic}\"",
		"notification.debate.turn":                "It's your turn in \"{topic}\"",
		"notification.debate.turnExpiring":        "Your turn is expiring soon in \"{topic}\"",
		"notification.debate.ended":               "Debate ended in \"{topic}\"",
		"notification.debate.walkover":            "Debate ended in \"{topic}\" — walkover",
		"notification.debate.seatOpened":          "A seat opened in \"{topic}\"",
		"notification.debate.replacementJoined":   "A replacement debater joined \"{topic}\"",
		"notification.debate.noReplacement":       "No replacement found for \"{topic}\". You win by walkover.",
		"notification.debate.expiredNoJoin":       "Your debate \"{topic}\" expired. No one joined.",
		"notification.draw.proposed":              "Your opponent proposed a draw in \"{topic}\"",
		"notification.extension.declined":         "Extension declined in \"{topic}\". The debate ended in a draw.",
		"notification.extension.accepted":         "Extension accepted in \"{topic}\". The debate continues!",
		"notification.extension.invite":           "Your opponent agreed to extend \"{topic}\". Accept to continue!",
		"notification.extension.expired":          "Extension expired in \"{topic}\". The debate ended in a draw.",
		"notification.moderation.hidden":          "A moderator hid {target}{context}.{note}",
		"notification.moderation.restored":        "A moderator restored {target}{context}.{note}",
		"notification.moderation.target.debate":   "your debate",
		"notification.moderation.target.turn":     "your turn #{turn}",
		"notification.moderation.target.comment":  "your comment",
		"notification.moderation.context":         " in \"{topic}\"",
		"notification.moderation.note":            " Moderator note: {note}",
		"notification.enforcement.action":         "A moderator issued an account {action} action.{note}",
		"notification.enforcement.restricted":     "A moderator restricted your account actions: {capabilities}.{expires}{note}",
		"notification.enforcement.expires":        " Effective until {expiresAt}.",
		"notification.enforcement.revoked":        "A moderator revoked your {action} action. Moderator note: {note}",
	},
	LocaleVI: {
		"error.server":              "Đã xảy ra lỗi.",
		"error.authRequired":        "Cần đăng nhập.",
		"error.mustSignIn":          "Bạn cần đăng nhập.",
		"error.verifyEmail":         "Vui lòng xác minh email để thực hiện hành động này.",
		"error.permissionDenied":    "Bạn không có quyền thực hiện hành động này.",
		"error.accountSuspended":    "Tài khoản này đang tạm thời bị đình chỉ.",
		"error.accountBanned":       "Tài khoản này đã bị cấm vĩnh viễn.",
		"error.invalidRequestBody":  "Nội dung yêu cầu không hợp lệ.",
		"error.invalidPage":         "Trang không hợp lệ.",
		"error.invalidPerPage":      "per_page không hợp lệ.",
		"error.validation":          "Yêu cầu không hợp lệ.",
		"error.debateNotFound":      "Tranh luận này không tồn tại hoặc đã bị gỡ bỏ.",
		"error.debateHiddenByBlock": "Tranh luận này bị ẩn bởi cài đặt an toàn của bạn.",
		"error.localeInvalid":       "locale phải là một trong: en, vi.",

		"notification.challenge.expired":          "Lời thách đấu của bạn cho \"{topic}\" đã hết hạn.",
		"notification.challenge.declined":         "Lời thách đấu của bạn cho \"{topic}\" đã bị từ chối.",
		"notification.challenge.accepted":         "Lời thách đấu của bạn cho \"{topic}\" đã được chấp nhận.",
		"notification.challenge.invited":          "@{username} đã mời bạn tham gia: \"{topic}\"",
		"notification.challenge.challenged":       "@{username} đã thách đấu bạn: \"{topic}\"",
		"notification.challenge.previousOpponent": "Đối thủ trước đây đã thách đấu bạn: \"{topic}\"",
		"notification.comment.reply":              "Có người đã trả lời phản hồi của bạn trong \"{topic}\"",
		"notification.debate.joined":              "Có người đã tham gia tranh luận \"{topic}\" của bạn",
		"notification.debate.turn":                "Đến lượt bạn trong \"{topic}\"",
		"notification.debate.turnExpiring":        "Lượt của bạn trong \"{topic}\" sắp hết hạn",
		"notification.debate.ended":               "Tranh luận trong \"{topic}\" đã kết thúc",
		"notification.debate.walkover":            "Tranh luận trong \"{topic}\" đã kết thúc — thắng do đối phương bỏ cuộc",
		"notification.debate.seatOpened":          "Một vị trí đã mở trong \"{topic}\"",
		"notification.debate.replacementJoined":   "Một người thay thế đã tham gia \"{topic}\"",
		"notification.debate.noReplacement":       "Không tìm được người thay thế cho \"{topic}\". Bạn thắng do đối phương bỏ cuộc.",
		"notification.debate.expiredNoJoin":       "Tranh luận \"{topic}\" của bạn đã hết hạn. Không có ai tham gia.",
		"notification.draw.proposed":              "Đối thủ đề nghị hòa trong \"{topic}\"",
		"notification.extension.declined":         "Gia hạn bị từ chối trong \"{topic}\". Tranh luận kết thúc với kết quả hòa.",
		"notification.extension.accepted":         "Gia hạn được chấp nhận trong \"{topic}\". Tranh luận tiếp tục!",
		"notification.extension.invite":           "Đối thủ đã đồng ý gia hạn \"{topic}\". Chấp nhận để tiếp tục!",
		"notification.extension.expired":          "Gia hạn trong \"{topic}\" đã hết hạn. Tranh luận kết thúc với kết quả hòa.",
		"notification.moderation.hidden":          "Kiểm duyệt viên đã ẩn {target}{context}.{note}",
		"notification.moderation.restored":        "Kiểm duyệt viên đã khôi phục {target}{context}.{note}",
		"notification.moderation.target.debate":   "tranh luận của bạn",
		"notification.moderation.target.turn":     "lượt #{turn} của bạn",
		"notification.moderation.target.comment":  "bình luận của bạn",
		"notification.moderation.context":         " trong \"{topic}\"",
		"notification.moderation.note":            " Ghi chú kiểm duyệt: {note}",
		"notification.enforcement.action":         "Kiểm duyệt viên đã áp dụng biện pháp {action} cho tài khoản.{note}",
		"notification.enforcement.restricted":     "Kiểm duyệt viên đã hạn chế các hành động tài khoản của bạn: {capabilities}.{expires}{note}",
		"notification.enforcement.expires":        " Có hiệu lực đến {expiresAt}.",
		"notification.enforcement.revoked":        "Kiểm duyệt viên đã thu hồi biện pháp {action} của bạn. Ghi chú kiểm duyệt: {note}",
	},
}
