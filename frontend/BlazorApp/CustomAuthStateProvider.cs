namespace BlazorApp {
	using System.Net;
	using System.Runtime.Serialization.Formatters.Binary;
	using System.Security.Claims;
	using System.Text.Json;
	using BitzArt.Blazor.Cookies;
	using Microsoft.AspNetCore.Authentication;
	using Microsoft.AspNetCore.Authentication.Cookies;
	using Microsoft.AspNetCore.Components.Authorization;
	using Microsoft.AspNetCore.Http;

	public class CustomAuthStateProvider : AuthenticationStateProvider {
		public override Task<AuthenticationState> GetAuthenticationStateAsync() {
			// TODO Get state from session
			return Task.FromResult(new AuthenticationState(new ClaimsPrincipal()));
		}
	}
}
