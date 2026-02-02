package handler

import (
	"net/http"

	"jd_material_push/internal/logic"
	"jd_material_push/internal/svc"
	"jd_material_push/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func SubmitMaterialBatchHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SubmitMaterialBatchRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewSubmitMaterialBatchLogic(r.Context(), svcCtx)
		resp, err := l.SubmitMaterialBatch(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
