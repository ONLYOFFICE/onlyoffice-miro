/**
 *
 * (c) Copyright Ascensio System SIA 2025
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
package controller

import (
	"html/template"
	"io"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/assets"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	echo "github.com/labstack/echo/v4"
)

type TemplateRenderer struct {
	templates *template.Template
	logger    service.Logger
}

func NewTemplateRenderer(logger service.Logger) (TemplateRenderer, error) {
	templates, err := template.ParseFS(assets.Views, "views/*.html")
	if err != nil {
		return TemplateRenderer{}, err
	}

	return TemplateRenderer{
		templates: templates,
		logger:    logger,
	}, nil
}

func (r *TemplateRenderer) Render(w io.Writer, name string, data any, c echo.Context) error {
	err := r.templates.ExecuteTemplate(w, name, data)
	if err != nil {
		r.logger.Error(c.Request().Context(), "Failed to render template", service.Fields{
			"template": name,
			"error":    err,
		})
	}

	return err
}
