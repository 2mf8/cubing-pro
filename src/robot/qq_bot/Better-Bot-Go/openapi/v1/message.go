package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/dto"
	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/errs"
	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/openapi"
	"github.com/tidwall/gjson"
)

// Message 拉取单条消息
func (o *openAPI) Message(ctx context.Context, channelID string, messageID string) (*dto.Message, error) {
	resp, err := o.request(ctx).
		SetResult(dto.Message{}).
		SetPathParam("channel_id", channelID).
		SetPathParam("message_id", messageID).
		Get(o.getURL(messageURI))
	if err != nil {
		return nil, err
	}

	// 兼容处理
	result := resp.Result().(*dto.Message)
	if result.ID == "" {
		body := gjson.Get(resp.String(), "message")
		if err := json.Unmarshal([]byte(body.String()), result); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// Messages 拉取消息列表
func (o *openAPI) Messages(ctx context.Context, channelID string, pager *dto.MessagesPager) ([]*dto.Message, error) {
	if pager == nil {
		return nil, errs.ErrPagerIsNil
	}
	resp, err := o.request(ctx).
		SetPathParam("channel_id", channelID).
		SetQueryParams(pager.QueryParams()).
		Get(o.getURL(messagesURI))
	if err != nil {
		return nil, err
	}

	messages := make([]*dto.Message, 0)
	if err := json.Unmarshal(resp.Body(), &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

// PostMessage 发消息
func (o *openAPI) PostMessage(ctx context.Context, channelID string, msg *dto.MessageToCreate) (*dto.Message, error) {
	resp, err := o.request(ctx).
		SetResult(dto.Message{}).
		SetPathParam("channel_id", channelID).
		SetBody(msg).
		Post(o.getURL(messagesURI))
	if err != nil {
		return nil, err
	}

	return resp.Result().(*dto.Message), nil
}

// 频道发 file_image
func (o *openAPI) PostFormFileImage(ctx context.Context, channelID string, m map[string]string, path string) (*dto.Message, error) {
	resp, err := o.request(ctx).
		SetHeader("Content-Type", "multipart/form-data").
		SetResult(dto.Message{}).
		SetPathParam("channel_id", channelID).
		SetFormData(m).
		SetFile("file_image", path).
		Post(o.getURL(messagesURI))
	if err != nil {
		return nil, err
	}

	return resp.Result().(*dto.Message), nil
}

// 频道发 file_image
func (o *openAPI) PostFormFileReaderImage(ctx context.Context, channelID string, m map[string]string, filename string, r io.Reader) (*dto.Message, error) {
	resp, err := o.request(ctx).
		SetHeader("Content-Type", "multipart/form-data").
		SetResult(dto.Message{}).
		SetPathParam("channel_id", channelID).
		SetFormData(m).
		SetFileReader("file_image", filename, r).
		Post(o.getURL(messagesURI))
	if err != nil {
		return nil, err
	}

	return resp.Result().(*dto.Message), nil
}

func (o *openAPI) PostGroupMessage(ctx context.Context, groupID string, msg *dto.GroupMessageToCreate) (*dto.GroupMsgResp, error) {
	resp, err := o.request(ctx).
		SetResult(dto.GroupMsgResp{}).
		SetPathParam("group_openid", groupID).
		SetBody(msg).
		Post(o.getURL(groupMessageUri))
	if err != nil {
		return nil, err
	}

	return resp.Result().(*dto.GroupMsgResp), nil
}

func (o *openAPI) PostC2CMessage(ctx context.Context, userId string, msg *dto.C2CMessageToCreate) (*dto.C2CMsgResp, error) {
	resp, err := o.request(ctx).
		SetResult(dto.C2CMsgResp{}).
		SetPathParam("openid", userId).
		SetBody(msg).
		Post(o.getURL(privateMessageUri))
	if err != nil {
		return nil, err
	}

	return resp.Result().(*dto.C2CMsgResp), nil
}

func (o *openAPI) PostC2CRichMediaMessage(ctx context.Context, userId string, msg *dto.C2CRichMediaMessageToCreate) (*dto.RichMediaMsgResp, error) {
	resp, err := o.request(ctx).
		SetResult(dto.RichMediaMsgResp{}).
		SetPathParam("openid", userId).
		SetBody(msg).
		Post(o.getURL(privateRichMediaMessageUri))
	if err != nil {
		return nil, err
	}

	return resp.Result().(*dto.RichMediaMsgResp), nil
}

func (o *openAPI) PostGroupRichMediaMessage(ctx context.Context, groupID string, msg *dto.GroupRichMediaMessageToCreate) (*dto.RichMediaMsgResp, error) {
	resp, err := o.request(ctx).
		SetResult(dto.RichMediaMsgResp{}).
		SetPathParam("group_openid", groupID).
		SetBody(msg).
		Post(o.getURL(groupRichMediaMessageUri))
	if err != nil {
		return nil, err
	}

	return resp.Result().(*dto.RichMediaMsgResp), nil
}

// PatchMessage 编辑消息
func (o *openAPI) PatchMessage(ctx context.Context,
	channelID string, messageID string, msg *dto.MessageToCreate) (*dto.Message, error) {
	resp, err := o.request(ctx).
		SetResult(dto.Message{}).
		SetPathParam("channel_id", channelID).
		SetPathParam("message_id", messageID).
		SetBody(msg).
		Patch(o.getURL(messageURI))
	if err != nil {
		return nil, err
	}

	return resp.Result().(*dto.Message), nil
}

// RetractMessage 撤回消息
func (o *openAPI) RetractMessage(ctx context.Context,
	channelID, msgID string, options ...openapi.RetractMessageOption) error {
	request := o.request(ctx).
		SetPathParam("channel_id", channelID).
		SetPathParam("message_id", string(msgID))
	for _, option := range options {
		if option == openapi.RetractMessageOptionHidetip {
			request = request.SetQueryParam("hidetip", "true")
		}
	}
	_, err := request.Delete(o.getURL(messageURI))
	return err
}

// DelC2CMessage 撤回C2C机器人2分钟内发出的消息
func (o *openAPI) DelC2CMessage(ctx context.Context,
	userID, msgID string, options ...openapi.RetractMessageOption) error {
	request := o.request(ctx).
		SetPathParam("openid", userID).
		SetPathParam("message_id", string(msgID))
	for _, option := range options {
		if option == openapi.RetractMessageOptionHidetip {
			request = request.SetQueryParam("hidetip", "true")
		}
	}
	_, err := request.Delete(o.getURL(privateBotMessageDelUri))
	return err
}

// DelGroupBotMessage 撤回机器人2分钟内发出的群聊消息
func (o *openAPI) DelGroupBotMessage(ctx context.Context,
	groupID, msgID string, options ...openapi.RetractMessageOption) error {
	request := o.request(ctx).
		SetPathParam("group_openid", groupID).
		SetPathParam("message_id", string(msgID))
	for _, option := range options {
		if option == openapi.RetractMessageOptionHidetip {
			request = request.SetQueryParam("hidetip", "true")
		}
	}
	_, err := request.Delete(o.getURL(groupBotMessageDelUri))
	return err
}

// PostSettingGuide 发送设置引导消息, atUserID为要at的用户
func (o *openAPI) PostSettingGuide(ctx context.Context,
	channelID string, atUserIDs []string) (*dto.Message, error) {
	var content string
	for _, userID := range atUserIDs {
		content += fmt.Sprintf("<@%s>", userID)
	}
	msg := &dto.SettingGuideToCreate{
		Content: content,
	}
	resp, err := o.request(ctx).
		SetResult(dto.Message{}).
		SetPathParam("channel_id", channelID).
		SetBody(msg).
		Post(o.getURL(settingGuideURI))
	if err != nil {
		return nil, err
	}
	return resp.Result().(*dto.Message), nil
}
