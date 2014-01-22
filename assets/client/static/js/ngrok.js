var ngrok = angular.module("ngrok", ["ngSanitize"]);

var hexRepr = function(bytes) {
    var buf = [];
    var ascii = [];
    for (var i=0; i<bytes.length; ++i) {
        var b = bytes[i];

        if (!(i%8) && i!=0) {
            buf.push("\t");
            buf.push.apply(buf, ascii)
            buf.push('\n');
            ascii = [];
        }

        if (b < 16) {
            buf.push("0");
        }

        if (b < 0x20 || b > 0x7e) {
            ascii.push('.');
        } else {
            ascii.push(String.fromCharCode(b));
        }

        buf.push(b.toString(16));
        buf.push(" ");
        ascii.push(" ");
    }

    if (ascii.length > 0) {
        var charsLeft = 8 - (ascii.length / 2);
        for (i=0; i<charsLeft; ++i) {
            buf.push("   ");
        }
        buf.push("\t");
        buf.push.apply(buf, ascii);
    }

    return buf.join("");
}

ngrok.factory("txnSvc", function() {
    var processBody = function(body, binary) {
        body.binary = binary;
        body.isForm = body.ContentType == "application/x-www-form-urlencoded";
        body.exists = body.Length > 0;
        body.hasError = !!body.Error;

        var syntaxClass = {
            "text/xml":               "xml",
            "application/xml":        "xml",
            "text/html":              "xml",
            "text/css":               "css",
            "application/json":       "json",
            "text/javascript":        "javascript",
            "application/javascript": "javascript",
        }[body.ContentType];

        // decode body
        if (binary) {
            body.Text = "";
        } else {
            body.Text = Base64.decode(body.Text).text;
        }

        // prettify
        var transform = {
            "xml": "xml",
            "json": "json"
        }[syntaxClass];

        if (!body.hasError && !!transform) {
            try {
                // vkbeautify does poorly at formatting html
                if (body.ContentType != "text/html") {
                    body.Text = vkbeautify[transform](body.Text);
                }
            } catch (e) {
            }
        }

        if (!!syntaxClass) {
            body.Text = hljs.highlight(syntaxClass, body.Text).value;
        } else {
            // highlight.js doesn't have a 'plaintext' syntax, so we'll just copy its escaping function.
            body.Text = body.Text.replace(/&/gm, '&amp;').replace(/</gm, '&lt;').replace(/>/gm, '&gt;');
        }
    };

    var processReq = function(req) {
        if (!req.RawBytes) {
            var decoded = Base64.decode(req.Raw);
            req.RawBytes = hexRepr(decoded.bytes);

            if (!req.Binary) {
                req.RawText = decoded.text;
            }
        }

        processBody(req.Body, req.Binary);
    };

    var processResp = function(resp) {
        resp.statusClass = {
            '2': "text-info",
            '3': "muted",
            '4': "text-warning",
            '5': "text-error"
        }[resp.Status[0]];

        if (!resp.RawBytes) {
            var decoded = Base64.decode(resp.Raw);
            resp.RawBytes = hexRepr(decoded.bytes);

            if (!resp.Binary) {
                resp.RawText = decoded.text;
            }
        }

        processBody(resp.Body, resp.Binary);
    };

    var processTxn = function(txn) {
        processReq(txn.Req);
        processResp(txn.Resp);
    };

    var preprocessTxn = function(txn) {
        var toFixed = function(value, precision) {
            var power = Math.pow(10, precision || 0);
            return String(Math.round(value * power) / power);
        }
        // parse nanosecond count
        var ns = txn.Duration;
        var ms = ns / (1000 * 1000);
        txn.Duration = ms;
        if (ms > 1000) {
            txn.Duration = toFixed(ms / 1000, 2) + "s";
        } else {
            txn.Duration = toFixed(ms, 2) + "ms";
        }
    };


    var active;
    var txns = window.data.Txns;
    txns.forEach(function(t) {
        preprocessTxn(t);
    });

    var activate = function(txn) {
        if (!txn.processed) {
            processTxn(txn);
            txn.processed = true;
        }
        active = txn;
    }

    if (txns.length > 0) {
        activate(txns[0]);
    }

    return {
        add: function(txnData) {
            txns.unshift(JSON.parse(txnData));
            preprocessTxn(txns[0]);
            if (!active) {
                activate(txns[0]);
            }
        },
        all: function() {
            return txns;
        },
        active: function(txn) {
            if (!txn) {
                return active;
            } else {
                activate(txn);
            }
        },
        isActive: function(txn) {
            return !!active && txn.Id == active.Id;
        }
    };
});

ngrok.directive({
    "keyval": function() {
        return {
            scope: {
                title: "@",
                tuples: "=",
            },
            replace: true,
            restrict: "E",
            template: "" +
            '<div ng-show="hasKeys()">' +
                '<h6>{{title}}</h6>' +
                '<table class="table params">' +
                    '<tr ng-repeat="(key, value) in tuples">' +
                        '<th>{{ key }}</th>' +
                        '<td>{{ value }}</td>' +
                    '</tr>' +
                '</table>' +
            '</div>',
            link: function($scope) {
                $scope.hasKeys = function() {
                    for (key in $scope.tuples) { return true; }
                    return false;
                };
            }
        };
    },

    "tabs": function() {
        return {
            scope: {
                "tabs": "@",
                "btn": "@",
                "onbtnclick": "&"
            },
            replace: true,
            template: '' +
            '<ul class="nav nav-pills">' +
                '<li ng-repeat="tab in tabNames" ng-class="{\'active\': isTab(tab)}">' +
                    '<a href="" ng-click="setTab(tab)">{{tab}}</a>' +
                '</li>' +
                '<li ng-show="!!btn" class="pull-right"> <button class="btn btn-primary" ng-click="onbtnclick()">{{btn}}</button></li>' +
            '</ul>',
            link: function postLink(scope, element, attrs) {
                scope.tabNames = attrs.tabs.split(",");
                scope.activeTab = scope.tabNames[0];
                scope.setTab = function(t) {
                    scope.activeTab = t;
                };
                scope.$parent.isTab = scope.isTab = function(t) {
                    return t == scope.activeTab;
                };
            },
        };
    },

    "body": function() {
        return {
            scope: {
                "body": "=",
                "binary": "="
            },
            template: '' +
            '<h6 ng-show="body.exists">' +
                '{{ body.Length }} bytes ' +
                '{{ body.RawContentType }}' +
            '</h6>' +
'' +
            '<div ng-show="!body.isForm && !body.binary">' +
                '<pre ng-show="body.exists"><code ng-bind-html="body.Text"></code></pre>' +
            '</div>' +
'' +
            '<div ng-show="body.isForm">' +
                '<keyval title="Form Params" tuples="body.Form">' +
            '</div>' +
            '<div ng-show="body.hasError" class="alert">' +
                '{{ body.Error }}' +
            '</div>',

            link: function($scope, $elem) {
                $scope.$watch(function() { return $scope.body; }, function() {
                    if ($scope.body && $scope.body.ErrorOffset > -1) {
                        var offset = $scope.body.ErrorOffset;

                        function textNodes(node) {
                            var textNodes = [];

                            function getTextNodes(node) {
                                if (node.nodeType == 3) {
                                    textNodes.push(node);
                                } else {
                                    for (var i = 0, len = node.childNodes.length; i < len; ++i) {
                                        getTextNodes(node.childNodes[i]);
                                    }
                                }
                            }

                            getTextNodes(node);
                            return textNodes;
                        }

                        var tNodes = textNodes($elem.find("code").get(0));
                        for (var i=0; i<tNodes.length; i++) {
                            offset -= tNodes[i].nodeValue.length;
                            if (offset < 0) {
                                $(tNodes[i]).parent().css("background-color", "orange");
                                break;
                            }
                        }
                    }
                });
            }
        };
    }
});

ngrok.controller({
    "HttpTxns": function($scope, txnSvc) {
        $scope.tunnels = window.data.UiState.Tunnels;
        $scope.txns = txnSvc.all();

        if (!!window.WebSocket) {
            var ws = new WebSocket("ws://" + location.host + "/_ws");
            ws.onopen = function() {
                console.log("connected websocket for real-time updates");
            };

            ws.onmessage = function(message) {
                $scope.$apply(function() {
                    txnSvc.add(message.data);
                });
            };

            ws.onerror = function(err) {
                console.log("Web socket error:")
                console.log(err);
            };

            ws.onclose = function(cls) {
                console.log("Web socket closed:" + cls);
            };
        }
    },

    "HttpRequest": function($scope, txnSvc) {
        $scope.replay = function() {
            $.ajax({
                type: "POST",
                url: "/http/in/replay",
                data: { txnid: txnSvc.active().Id }
            });
        }
        var setReq = function() {
            var txn = txnSvc.active();
            if (!!txn && txn.Req) {
                $scope.Req = txnSvc.active().Req;
            } else {
                $scope.Req = null;
            }
        };
        $scope.$watch(function() { return txnSvc.active() }, setReq);
    },

    "HttpResponse": function($scope, $element, txnSvc) {
        var setResp = function() {
            var txn = txnSvc.active();
            if (!!txn && txn.Resp) {
                $scope.Resp = txnSvc.active().Resp;
            } else {
                $scope.Resp = null;
            }
        };
        $scope.$watch(function() { return txnSvc.active() }, setResp);
    },

    "TxnNavItem": function($scope, txnSvc) {
        $scope.isActive = function() { return txnSvc.isActive($scope.txn); }
        $scope.makeActive = function() {
            txnSvc.active($scope.txn);
        };
    },

    "HttpTxn": function($scope, txnSvc, $timeout) {
        var setTxn = function() {
            $scope.Txn = txnSvc.active();
        };

        $scope.ISO8601 = function(ts) {
            if (!!ts) {
                return new Date(ts * 1000).toISOString();
            }
        };

        $scope.TimeFormat = function(ts) {
            if (!!ts) {
                return $.timeago($scope.ISO8601(ts));
            }
        };

        $scope.$watch(function() { return txnSvc.active() }, setTxn);

        // this causes angular to update the timestamps
        setInterval(function() { $scope.$apply(function() {}); }, 30000);
    },
});
