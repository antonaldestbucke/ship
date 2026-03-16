async function getStars(): Promise<number | null> {
  try {
    const res = await fetch(
      "https://api.github.com/repos/basilysf1709/ship",
      { next: { revalidate: 3600 } }
    );
    if (!res.ok) return null;
    const data = await res.json();
    return data.stargazers_count ?? null;
  } catch {
    return null;
  }
}

export async function GitHubStars() {
  const stars = await getStars();
  if (stars === null) return null;
  return <span>{stars}</span>;
}
