var ngrok = angular.module("ngrok", []);

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
                    '<a ng-click="setTab(tab)" href="#">{{tab}}</a>' +
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
                "body": "="
            },
            template: '' +
            '<h6 ng-show="hasBody">' +
                '{{ Body.Length }} bytes ' +
                '{{ Body.RawContentType }}' +
            '</h6>' +
'' +
            '<div ng-show="!isForm">' +
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
                    body.Text = vkbeautify[transform](body.Text);
                    } catch (e) {
                    }
                }

                $scope.Body = body;
            }
        };
    }
});

ngrok.controller({
    "HttpTxns": function($scope) {
        $scope.txns = window.txns;

        if (!!window.WebSocket) {
            var ws = new WebSocket("ws://localhost:4040/_ws");
            ws.onopen = function() {
                console.log("connected websocket for real-time updates");
            };

            ws.onmessage = function(message) {
                $scope.$apply(function() {
                    $scope.txns.unshift(JSON.parse(message.data));
                });

                /*
                $("pre code").each(function(i, e) {
                    hljs.highlightBlock(e)
                });
                */
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
    },

    "HttpResponse": function($scope) {
        $scope.Resp = $scope.txn.Resp;
        $scope.statusClass = {
            '2': "text-info",
            '3': "muted",
            '4': "text-warning",
            '5': "text-error"
        }[$scope.Resp.Status[0]];
    }
});
