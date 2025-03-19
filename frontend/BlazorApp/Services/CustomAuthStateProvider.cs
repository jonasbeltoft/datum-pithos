namespace BlazorApp.Services {
	using System.Net;
	using System.Runtime.Serialization.Formatters.Binary;
	using System.Security.Claims;
	using System.Text.Json;
	using Microsoft.AspNetCore.Authentication;
	using Microsoft.AspNetCore.Authentication.Cookies;
	using Microsoft.AspNetCore.Components.Authorization;
	using Microsoft.AspNetCore.Http;

	public class CustomAuthStateProvider : AuthenticationStateProvider {
		public CustomAuthStateProvider() {
		}

		public override Task<AuthenticationState> GetAuthenticationStateAsync() {
			// TODO Get state from session
			return Task.FromResult(new AuthenticationState(new ClaimsPrincipal()));
		}
	}
}
