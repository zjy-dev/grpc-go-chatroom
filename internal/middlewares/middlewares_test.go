//go:build unit_test

package middlewares

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockHandler is a mock implementation of grpc.UnaryHandler
type MockHandler struct {
	mock.Mock
}

func (m *MockHandler) Handle(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

// MockAuthFunc is a mock implementation of the AuthFunc
type MockAuthFunc struct {
	mock.Mock
}

// Authenticate implements the AuthFunc interface
func (m *MockAuthFunc) Authenticate(ctx context.Context) (context.Context, error) {
	args := m.Called(ctx)
	return args.Get(0).(context.Context), args.Error(1)
}

func TestUnaryServerInterceptorWithBypassMethods(t *testing.T) {
	require := require.New(t)
	mockHandler := new(MockHandler)
	mockAuthFunc := new(MockAuthFunc)

	tests := []struct {
		name           string
		method         string
		bypassMethods  []string
		authFuncResult error
		expectAuthCall bool
		expectError    bool
	}{
		{
			name:           "MethodInBypassList",
			method:         "/service.Service/BypassMethod",
			bypassMethods:  []string{"/service.Service/BypassMethod"},
			authFuncResult: nil,
			expectAuthCall: false,
			expectError:    false,
		},
		{
			name:           "MethodNotInBypassList",
			method:         "/service.Service/NormalMethod",
			bypassMethods:  []string{"/service.Service/BypassMethod"},
			authFuncResult: nil,
			expectAuthCall: true,
			expectError:    false,
		},
		{
			name:           "AuthFuncReturnsError",
			method:         "/service.Service/NormalMethod",
			bypassMethods:  []string{},
			authFuncResult: status.Errorf(codes.Unauthenticated, "unauthenticated"),
			expectAuthCall: true,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := UnaryServerAuthInterceptorWithBypassMethods(mockAuthFunc.Authenticate, tt.bypassMethods...)

			ctx := context.Background()
			req := struct{}{}
			info := &grpc.UnaryServerInfo{
				FullMethod: tt.method,
			}

			if !tt.expectAuthCall || tt.authFuncResult == nil {
				mockHandler.On("Handle", ctx, req).Return(struct{}{}, nil).Once()
			}

			if tt.expectAuthCall {
				mockAuthFunc.On("Authenticate", ctx).Return(ctx, tt.authFuncResult).Once()
			}

			resp, err := interceptor(ctx, req, info, mockHandler.Handle)

			if tt.expectAuthCall {
				mockAuthFunc.AssertCalled(t, "Authenticate", ctx)
				if tt.authFuncResult != nil {
					require.Error(err)
					require.Equal(tt.authFuncResult, err)
				} else {
					require.NoError(err)
					require.Equal(struct{}{}, resp)
				}
			} else {
				mockAuthFunc.AssertNotCalled(t, "Authenticate", ctx)
				require.NoError(err)
				require.Equal(struct{}{}, resp)
			}

			mockHandler.AssertExpectations(t)
			mockAuthFunc.AssertExpectations(t)
		})
	}
}
