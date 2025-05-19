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

/* eslint-disable react/button-has-type */
import React, { forwardRef } from 'react';

import '@components/button.css';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  name: string;
  variant?: 'primary' | 'default';
}

const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      id,
      name,
      disabled,
      value,
      className = '',
      variant = 'default',
      onClick,
      type,
      ...props
    },
    ref
  ) => {
    const realId = id || Math.random().toString(36).substring(2, 9);

    const handleClick = (e: React.MouseEvent<HTMLButtonElement>) => {
      if (onClick) onClick(e);
    };

    return (
      <button
        id={realId}
        ref={ref}
        disabled={disabled}
        onClick={handleClick}
        className={`generic-button ${variant === 'primary' ? 'primary' : ''} ${className}`}
        type={type || 'button'}
        {...props}
      >
        {name}
      </button>
    );
  }
);

Button.displayName = 'Button';

export default Button;
