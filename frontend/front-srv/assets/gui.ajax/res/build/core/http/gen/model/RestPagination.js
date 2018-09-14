/**
 * Pydio Cells Rest API
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * OpenAPI spec version: 1.0
 * 
 *
 * NOTE: This class is auto generated by the swagger code generator program.
 * https://github.com/swagger-api/swagger-codegen.git
 * Do not edit the class manually.
 *
 */

'use strict';

exports.__esModule = true;

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError('Cannot call a class as a function'); } }

var _ApiClient = require('../ApiClient');

var _ApiClient2 = _interopRequireDefault(_ApiClient);

/**
* The RestPagination model module.
* @module model/RestPagination
* @version 1.0
*/

var RestPagination = (function () {
    /**
    * Constructs a new <code>RestPagination</code>.
    * @alias module:model/RestPagination
    * @class
    */

    function RestPagination() {
        _classCallCheck(this, RestPagination);

        this.Limit = undefined;
        this.CurrentOffset = undefined;
        this.Total = undefined;
        this.CurrentPage = undefined;
        this.TotalPages = undefined;
        this.NextOffset = undefined;
        this.PrevOffset = undefined;
    }

    /**
    * Constructs a <code>RestPagination</code> from a plain JavaScript object, optionally creating a new instance.
    * Copies all relevant properties from <code>data</code> to <code>obj</code> if supplied or a new instance if not.
    * @param {Object} data The plain JavaScript object bearing properties of interest.
    * @param {module:model/RestPagination} obj Optional instance to populate.
    * @return {module:model/RestPagination} The populated <code>RestPagination</code> instance.
    */

    RestPagination.constructFromObject = function constructFromObject(data, obj) {
        if (data) {
            obj = obj || new RestPagination();

            if (data.hasOwnProperty('Limit')) {
                obj['Limit'] = _ApiClient2['default'].convertToType(data['Limit'], 'Number');
            }
            if (data.hasOwnProperty('CurrentOffset')) {
                obj['CurrentOffset'] = _ApiClient2['default'].convertToType(data['CurrentOffset'], 'Number');
            }
            if (data.hasOwnProperty('Total')) {
                obj['Total'] = _ApiClient2['default'].convertToType(data['Total'], 'Number');
            }
            if (data.hasOwnProperty('CurrentPage')) {
                obj['CurrentPage'] = _ApiClient2['default'].convertToType(data['CurrentPage'], 'Number');
            }
            if (data.hasOwnProperty('TotalPages')) {
                obj['TotalPages'] = _ApiClient2['default'].convertToType(data['TotalPages'], 'Number');
            }
            if (data.hasOwnProperty('NextOffset')) {
                obj['NextOffset'] = _ApiClient2['default'].convertToType(data['NextOffset'], 'Number');
            }
            if (data.hasOwnProperty('PrevOffset')) {
                obj['PrevOffset'] = _ApiClient2['default'].convertToType(data['PrevOffset'], 'Number');
            }
        }
        return obj;
    };

    /**
    * @member {Number} Limit
    */
    return RestPagination;
})();

exports['default'] = RestPagination;
module.exports = exports['default'];

/**
* @member {Number} CurrentOffset
*/

/**
* @member {Number} Total
*/

/**
* @member {Number} CurrentPage
*/

/**
* @member {Number} TotalPages
*/

/**
* @member {Number} NextOffset
*/

/**
* @member {Number} PrevOffset
*/