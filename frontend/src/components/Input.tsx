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

import React, { forwardRef } from 'react';

import { sanitizeFormInput } from '@utils/sanitizer';

import '@components/input.css';

interface FormInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  helperText?: string;
  fullWidth?: boolean;
  disabled?: boolean;
  required?: boolean;
  enableXssSanitization?: boolean;
}

const FormInput = forwardRef<HTMLInputElement, FormInputProps>(
  (
    {
      id,
      label,
      error,
      helperText,
      disabled = false,
      required = false,
      enableXssSanitization = true,
      value,
      className = '',
      onChange,
      ...props
    },
    ref
  ) => {
    const realId = id || Math.random().toString(36).substring(2, 9);

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
      if (onChange) {
        if (enableXssSanitization) {
          const sanitizedValue = sanitizeFormInput(e.target.value);
          const sanitizedEvent = {
            ...e,
            target: {
              ...e.target,
              value: sanitizedValue,
            },
          } as React.ChangeEvent<HTMLInputElement>;
          onChange(sanitizedEvent);
        } else {
          onChange(e);
        }
      }
    };

    return (
      <div className="form-input">
        {label && (
          <label htmlFor={realId} className="form-input__label">
            {label}
            {required && <span className="form-input__label_required">*</span>}
          </label>
        )}
        <div className="form-input__container">
          <input
            id={realId}
            ref={ref}
            className={`
            form-input__field
            ${error ? 'form-input__field--error' : ''}
            ${disabled ? 'form-input__field--disabled' : ''}
            ${props.readOnly ? 'form-input__field--readonly' : ''}
            ${className}
          `}
            value={value}
            onChange={handleChange}
            disabled={disabled}
            required={required}
            {...props}
          />
        </div>
        {error && (
          <p className="form-input__helper-text form-input__helper-text_error">
            {error}
          </p>
        )}
        {!error && helperText && (
          <p className="form-input__helper-text form-input__helper-text_normal">
            {helperText}
          </p>
        )}
      </div>
    );
  }
);

FormInput.displayName = 'FormInput';

export default FormInput;
