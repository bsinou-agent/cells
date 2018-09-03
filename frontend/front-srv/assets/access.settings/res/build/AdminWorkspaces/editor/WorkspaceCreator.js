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

'use strict';

Object.defineProperty(exports, '__esModule', {
    value: true
});

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var _react = require('react');

var _react2 = _interopRequireDefault(_react);

var _modelWorkspace = require('../model/Workspace');

var _modelWorkspace2 = _interopRequireDefault(_modelWorkspace);

var _FeaturesListWizard = require('./FeaturesListWizard');

var _FeaturesListWizard2 = _interopRequireDefault(_FeaturesListWizard);

var _FeaturesStepper = require('./FeaturesStepper');

var _FeaturesStepper2 = _interopRequireDefault(_FeaturesStepper);

var _TplFieldsChooser = require('./TplFieldsChooser');

var _TplFieldsChooser2 = _interopRequireDefault(_TplFieldsChooser);

exports['default'] = _react2['default'].createClass({
    displayName: 'WorkspaceCreator',

    mixins: [AdminComponents.MessagesConsumerMixin],

    propTypes: {
        type: _react2['default'].PropTypes.oneOf(['template', 'workspace']),
        save: _react2['default'].PropTypes.func,
        closeEditor: _react2['default'].PropTypes.func,
        className: _react2['default'].PropTypes.string
    },

    getInitialState: function getInitialState() {
        return {
            newName: this.context.getMessage(this.props.type == 'workspace' ? 'ws.6' : 'ws.7'),
            edit: 'driver',
            selectedDriver: this.props.type === 'workspace' ? 'gateway' : 'datasource',
            values: { template: {}, general: {}, driver: {} },
            templateSelectedFields: [],
            fieldStates: null
        };
    },

    isDirty: function isDirty() {
        return false;
    },

    componentDidMount: function componentDidMount() {
        if (!_modelWorkspace2['default'].DRIVERS) {
            _modelWorkspace2['default'].loadAvailableDrivers((function () {
                this.setState({ driversLoaded: true });
            }).bind(this));
        } else {
            this.setState({ driversLoaded: true });
        }
    },

    selectionChange: function selectionChange(editMeta, driver, template) {
        if (driver || template == '0') {
            if (driver && this.state.selectedDriver && this.state.selectedDriver != driver) {
                this.setState({ templateSelectedFields: [] });
            }
            this.setState({
                edit: editMeta,
                selectedDriver: driver,
                selectedTemplate: null,
                fieldStates: null
            });
        } else if (template) {
            this.setState({
                edit: editMeta,
                selectedDriver: null,
                selectedTemplate: template,
                fieldStates: null
            });
        } else {
            this.setState({
                edit: editMeta,
                fieldStates: null
            });
        }
    },

    onFormParameterChange: function onFormParameterChange(paramName, newValue, oldValue) {

        if (this.state.selectedDriver === 'fs' && paramName === 'PATH') {
            FuncUtils.bufferCallback('validate-parameter', 1000, (function () {
                this.validateDriverParameter(paramName, newValue);
            }).bind(this));
        }
    },

    validateDriverParameter: function validateDriverParameter(paramName, paramValue) {
        var fieldStates = this.state.fieldStates || {};
        fieldStates[paramName] = { msg: 'Validating value ... ' };
        this.setState({ fieldStates: fieldStates });

        PydioApi.getClient().request({
            get_action: 'validate_driver_field',
            name: paramName,
            value: paramValue,
            driver: this.state.selectedDriver
        }, (function (transport) {
            var r = transport.responseJSON;
            if (r.error) {
                fieldStates[paramName] = { error: r.error };
            } else if (r.msg) {
                fieldStates[paramName] = { msg: r.msg };
            } else if (fieldStates[paramName]) {
                delete fieldStates[paramName];
            }
            this.setState({ fieldStates: fieldStates });
        }).bind(this));
    },

    onFormChange: function onFormChange(newValues) {
        if (newValues['DISPLAY']) this.setState({ newName: newValues['DISPLAY'] });else if (newValues['DISPLAY']) this.setState({ newName: newValues['DISPLAY'] });
        var allValues = this.state.values;
        allValues[this.state.edit] = newValues;
        this.setState({ values: allValues });
    },

    save: function save() {
        this.props.save(this.props.type, this.state);
    },

    toggleTemplateSelectedField: function toggleTemplateSelectedField(name, value) {
        var selected = this.state.templateSelectedFields;
        if (value && selected.indexOf(name) == -1) {
            selected.push(name);
        } else if (!value && selected.indexOf(name) !== -1) {
            selected = LangUtils.arrayWithout(selected, selected.indexOf(name));
        }
        this.setState({ templateSelectedFields: selected });
    },

    updateValidStatus: function updateValidStatus(newStatus) {
        var validRecord = this.state.valid || {};
        validRecord[this.state.edit] = newStatus;
        this.setState({ valid: validRecord });
    },

    render: function render() {
        var _this = this;

        var editor = undefined,
            rightFill = false,
            additionalFeatureComponents = undefined;
        if (this.state.driversLoaded) {
            var formDefs = [],
                formValues = this.state.values[this.state.edit];
            var params = undefined;
            if (this.state.edit == 'general') {
                params = _modelWorkspace2['default'].DRIVERS.get('fs').params;
            } else if (this.state.selectedDriver) {
                params = _modelWorkspace2['default'].DRIVERS.get(this.state.selectedDriver).params;
            }
            if (params) {
                editor = _modelWorkspace2['default'].buildEditorStatic(params, formDefs, formValues, this.state.selectedTemplate && this.state.edit == 'general' ? 'mixed' : this.state.edit, this.props.type == 'template');
                if (this.state.fieldStates) {
                    (function () {
                        var fieldStates = _this.state.fieldStates;
                        formDefs.map(function (f) {
                            if (fieldStates[f.name] && fieldStates[f.name]['error']) {
                                f.errorText = fieldStates[f.name]['error'];
                            } else if (f.errorText) {
                                delete f['errorText'];
                            }
                            if (fieldStates[f.name] && fieldStates[f.name]['msg']) {
                                // Replace orginal description
                                f.warningText = fieldStates[f.name]['msg'];
                            } else {
                                delete f['warningText'];
                            }
                        });
                    })();
                }
                if (!formDefs.length) {
                    editor = _react2['default'].createElement(
                        'div',
                        null,
                        pydio.MessageHash['ajxp_admin.ws.68']
                    );
                } else {
                    editor = _react2['default'].createElement(PydioForm.FormPanel, {
                        parameters: formDefs,
                        values: formValues,
                        className: 'full-width',
                        onChange: this.onFormChange,
                        onParameterChange: this.onFormParameterChange,
                        onValidStatusChange: this.updateValidStatus,
                        depth: -2
                    });
                }
            }
        }

        var currentValid = true;
        if (this.state.valid) {
            LangUtils.objectValues(this.state.valid).map(function (v) {
                currentValid = currentValid && v;
            });
        }
        if (this.state.fieldStates) {
            LangUtils.objectValues(this.state.fieldStates).map(function (v) {
                if (v.error) currentValid = false;
            });
        }

        var leftNav = _react2['default'].createElement(_FeaturesStepper2['default'], {
            onSelectionChange: this.selectionChange,
            driversLoaded: this.state.driversLoaded,
            wizardType: this.props.type,
            formIsValid: currentValid,
            save: this.save,
            close: this.props.closeEditor,
            additionalComponent: additionalFeatureComponents
        });

        return _react2['default'].createElement(
            PydioComponents.PaperEditorLayout,
            {
                title: this.state.newName,
                leftNav: leftNav,
                contentFill: rightFill
            },
            editor
        );
    }

});
module.exports = exports['default'];