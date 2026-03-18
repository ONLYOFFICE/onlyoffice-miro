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

import React from 'react';

import '@components/spinner.css';

interface SpinnerProps {
  size?: 'small' | 'medium' | 'large';
  className?: string;
  style?: React.CSSProperties;
  variant?: 'default' | 'blue' | 'colorful';
}

const Spinner: React.FC<SpinnerProps> = ({
  size = 'medium',
  className = '',
  style = {},
  variant = 'default',
}) => {
  const defaultStyle = {
    margin: '16px',
    ...style,
  };

  const variantClass = variant !== 'default' ? `spinner_${variant}` : '';

  return (
    <div
      className={`spinner spinner_${size} ${variantClass} ${className}`}
      style={defaultStyle}
    >
      <div className="spinner__circle">
        <div className="spinner__circle_gradient" />
        <div className="spinner__circle_inner" />
      </div>
    </div>
  );
};

Spinner.displayName = 'Spinner';

export default Spinner;
