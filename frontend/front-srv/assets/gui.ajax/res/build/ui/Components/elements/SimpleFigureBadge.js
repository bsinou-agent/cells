/*
 * Copyright 2007-2017 Charles du Jeu - Abstrium SAS <team (at) pyd.io>
 * This file is part of Pydio.
 *
 * Pydio is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pydio is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Pydio.  If not, see <http://www.gnu.org/licenses/>.
 *
 * The latest code can be found at <https://pydio.com>.
 */

/**
 * Simple MuiPaper with a figure and a legend
 */
'use strict';

exports.__esModule = true;
exports['default'] = React.createClass({
    displayName: 'SimpleFigureBadge',

    propTypes: {
        colorIndicator: React.PropTypes.string,
        figure: React.PropTypes.number.isRequired,
        legend: React.PropTypes.string
    },

    getDefaultProps: function getDefaultProps() {
        return {
            colorIndicator: ''
        };
    },

    render: function render() {
        return React.createElement(
            ReactMUI.Paper,
            { style: { display: 'inline-block', marginLeft: 16 } },
            React.createElement(
                'div',
                { className: 'figure-badge', style: this.props.colorIndicator ? { borderLeftColor: this.props.colorIndicator } : {} },
                React.createElement(
                    'div',
                    { className: 'figure' },
                    this.props.figure
                ),
                React.createElement(
                    'div',
                    { className: 'legend' },
                    this.props.legend
                )
            )
        );
    }
});
module.exports = exports['default'];
