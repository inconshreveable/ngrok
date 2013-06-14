var ngrok = angular.module("ngrok", []);

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
            '<div ng-show="hasKeys">' +
                '<h6>{{title}}</h6>' +
                '<table class="table params">' +
                    '<tr ng-repeat="(key, value) in tuples">' +
                        '<th>{{ key }}</th>' +
                        '<td>{{ value }}</td>' +
                    '</tr>' +
                '</table>' +
            '</div>',
            link: function($scope) {
                $scope.hasKeys = false;
                for (key in $scope.tuples) { $scope.hasKeys = true; break; }
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

    "body": function($timeout) {
        return {
            scope: {
                "body": "=",
                "binary": "="
            },
            template: '' +
            '<h6 ng-show="hasBody">' +
                '{{ Body.Length }} bytes ' +
                '{{ Body.RawContentType }}' +
            '</h6>' +
'' +
            '<div ng-show="!isForm && !binary">' +
                '<pre ng-show="hasBody"><code ng-class="syntaxClass">{{ Body.Text }}</code></pre>' +
            '</div>' +
'' +
            '<div ng-show="isForm">' +
                '<keyval title="Form Params" tuples="Body.Form">' +
            '</div>' +
            '<div ng-show="hasError" class="alert">' +
                '{{ Body.Error }}' +
            '</div>',

            controller: function($scope) {
                var body = $scope.body;
                if ($scope.binary) {
                    body.Text = "";
                } else {
                    body.Text = Base64.decode(body.Text).text;
                }
                $scope.isForm = (body.ContentType == "application/x-www-form-urlencoded");
                $scope.hasBody = (body.Length > 0);
                $scope.hasError = !!body.Error;
                $scope.syntaxClass = {
                    "text/xml":               "xml",
                    "application/xml":        "xml",
                    "text/html":              "xml",
                    "text/css":               "css",
                    "application/json":       "json",
                    "text/javascript":        "javascript",
                    "application/javascript": "javascript",
                }[body.ContentType];

                var transform = {
                    "xml": "xml",
                    "json": "json"
                }[$scope.syntaxClass];

                if (!$scope.hasError && !!transform) {
                    try {
                        // vkbeautify does poorly at formatting html
                        if (body.ContentType != "text/html") {
                            body.Text = vkbeautify[transform](body.Text);
                        }
                    } catch (e) {
                    }
                }

                $scope.Body = body;
            },

            link: function($scope, $elem) {
                $timeout(function() {
                    $code = $elem.find("code").get(0);
                    hljs.highlightBlock($code);

                    if ($scope.Body.ErrorOffset > -1) {
                        var offset = $scope.Body.ErrorOffset;

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
    "HttpTxns": function($scope) {
        $scope.publicUrl = window.data.UiState.Url;
        $scope.txns = window.data.Txns;

        if (!!window.WebSocket) {
            var ws = new WebSocket("ws://localhost:4040/_ws");
            ws.onopen = function() {
                console.log("connected websocket for real-time updates");
            };

            ws.onmessage = function(message) {
                $scope.$apply(function() {
                    $scope.txns.unshift(JSON.parse(message.data));
                });
            };
            
            ws.onerror = function(err) {
                console.log("Web socket error:" + err);
            };

            ws.onclose = function(cls) {
                console.log("Web socket closed:" + cls);
            };
        }
    },

    "HttpRequest": function($scope) {
        $scope.Req = $scope.txn.Req;
        $scope.replay = function() {
            $.ajax({
                type: "POST",
                url: "/http/in/replay",
                data: { txnid: $scope.txn.Id }
            });
        }

        var decoded = Base64.decode($scope.Req.Raw);
        $scope.Req.RawBytes = hexRepr(decoded.bytes);

        if (!$scope.Req.Binary) {
            $scope.Req.RawText = decoded.text;
        }
    },

    "HttpResponse": function($scope) {
        $scope.Resp = $scope.txn.Resp;
        $scope.statusClass = {
            '2': "text-info",
            '3': "muted",
            '4': "text-warning",
            '5': "text-error"
        }[$scope.Resp.Status[0]];

        var decoded = Base64.decode($scope.Resp.Raw);
        $scope.Resp.RawBytes = hexRepr(decoded.bytes);

        if (!$scope.Resp.Binary) {
            $scope.Resp.RawText = decoded.text;
        }
    }
});
