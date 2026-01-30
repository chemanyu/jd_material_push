// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"fmt"
	"os"
	"time"

	"jd_material_push/internal/svc"
	"jd_material_push/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFilesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFilesLogic {
	return &GetFilesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFilesLogic) GetFiles(req *types.GetFilesRequest) (resp *types.GetFilesResponse, err error) {
	resp = &types.GetFilesResponse{
		Code:    200,
		Message: "success",
		Data:    make([]types.FileInfo, 0),
	}

	// 检查路径是否为空
	if req.Path == "" {
		resp.Code = 400
		resp.Message = "path parameter is required"
		return resp, nil
	}

	// 检查路径是否存在
	fileInfo, err := os.Stat(req.Path)
	if err != nil {
		if os.IsNotExist(err) {
			resp.Code = 404
			resp.Message = fmt.Sprintf("path not found: %s", req.Path)
			return resp, nil
		}
		resp.Code = 500
		resp.Message = fmt.Sprintf("failed to access path: %v", err)
		return resp, nil
	}

	// 检查是否是目录
	if !fileInfo.IsDir() {
		resp.Code = 400
		resp.Message = "path is not a directory"
		return resp, nil
	}

	// 读取目录中的所有文件
	entries, err := os.ReadDir(req.Path)
	if err != nil {
		resp.Code = 500
		resp.Message = fmt.Sprintf("failed to read directory: %v", err)
		return resp, nil
	}

	// 遍历文件列表
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			l.Logger.Errorf("failed to get file info for %s: %v", entry.Name(), err)
			continue
		}

		fullPath := fmt.Sprintf("%s/%s", req.Path, entry.Name())
		fileData := types.FileInfo{
			Name:    entry.Name(),
			Path:    fullPath,
			Size:    info.Size(),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime().Format(time.RFC3339),
		}
		resp.Data = append(resp.Data, fileData)
	}

	return resp, nil
}
