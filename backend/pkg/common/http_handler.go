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
package common

import (
	echo "github.com/labstack/echo/v4"
)

type Handler interface {
	Handlers() map[HTTPMethod]echo.HandlerFunc
}

type BaseHandler struct{}

func (h *BaseHandler) Handlers() map[HTTPMethod]echo.HandlerFunc {
	return nil
}

func NewHandler(handlers map[HTTPMethod]echo.HandlerFunc) Handler {
	return &handler{handlers: handlers}
}

type handler struct {
	handlers map[HTTPMethod]echo.HandlerFunc
}

func (h *handler) Handlers() map[HTTPMethod]echo.HandlerFunc {
	return h.handlers
}

type ErrorResponse struct {
	Error string `json:"error"`
}
