package api

import (
	"ALLinSSL/backend/internal/cert"
	"ALLinSSL/backend/public"
	"archive/zip"
	"bytes"
	"github.com/gin-gonic/gin"
	"strings"
)

func GetCertList(c *gin.Context) {
	var form struct {
		Search string `form:"search"`
		Page   int64  `form:"p"`
		Limit  int64  `form:"limit"`
	}
	err := c.Bind(&form)
	if err != nil {
		public.FailMsg(c, err.Error())
		return
	}
	certList, count, err := cert.GetList(form.Search, form.Page, form.Limit)
	if err != nil {
		public.FailMsg(c, err.Error())
		return
	}
	public.SuccessData(c, certList, count)
	return
}

func UploadCert(c *gin.Context) {
	var form struct {
		Key  string `form:"key"`
		Cert string `form:"cert"`
	}
	err := c.Bind(&form)
	if err != nil {
		public.FailMsg(c, err.Error())
		return
	}
	form.Key = strings.TrimSpace(form.Key)
	form.Cert = strings.TrimSpace(form.Cert)
	
	if form.Key == "" {
		public.FailMsg(c, "名称不能为空")
		return
	}
	if form.Cert == "" {
		public.FailMsg(c, "类型不能为空")
		return
	}
	sha256, err := cert.UploadCert(form.Key, form.Cert)
	if err != nil {
		public.FailMsg(c, err.Error())
		return
	}
	public.SuccessData(c, sha256, 0)
	return
}

func DelCert(c *gin.Context) {
	var form struct {
		ID string `form:"id"`
	}
	err := c.Bind(&form)
	if err != nil {
		public.FailMsg(c, err.Error())
		return
	}
	if form.ID == "" {
		public.FailMsg(c, "ID不能为空")
		return
	}
	err = cert.DelCert(form.ID)
	if err != nil {
		public.FailMsg(c, err.Error())
		return
	}
	public.SuccessMsg(c, "删除成功")
	return
}

func DownloadCert(c *gin.Context) {
	ID := c.Query("id")
	
	if ID == "" {
		public.FailMsg(c, "ID不能为空")
		return
	}
	certData, err := cert.GetCert(ID)
	if err != nil {
		public.FailMsg(c, err.Error())
		return
	}
	
	// 构建 zip 包（内存中）
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	
	for filename, content := range certData {
		if filename == "cert" || filename == "key" {
			writer, err := zipWriter.Create(filename + ".pem")
			if err != nil {
				public.FailMsg(c, err.Error())
				return
			}
			_, err = writer.Write([]byte(content))
			if err != nil {
				public.FailMsg(c, err.Error())
				return
			}
		}
	}
	// 关闭 zipWriter
	if err := zipWriter.Close(); err != nil {
		public.FailMsg(c, err.Error())
		return
	}
	// 设置响应头
	
	zipName := strings.ReplaceAll(certData["domains"], ".", "_")
	zipName = strings.ReplaceAll(zipName, ",", "-")
	
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename="+zipName+".zip")
	c.Data(200, "application/zip", buf.Bytes())
	return
}
