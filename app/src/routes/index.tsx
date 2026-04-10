import { Item, ItemContent, ItemDescription, ItemTitle } from '#/components/ui/item'
import { createFileRoute, Link } from '@tanstack/react-router'


export const Route = createFileRoute('/')({ component: MainPage })

function MainPage() {
    return <main className="page-wrap px-4 pb-8 pt-14">
        <section className="island-shell  max-w-2xl rounded-2xl">
            <h1 className="display-title mb-2 text-5xl sm:text-5xl">Utilitools</h1>
            <p className="mb-6 text-sm opacity-80">A collection of useful tools to help you with your daily tasks.</p>
        </section>

        <section className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Link to="/shorten-url">
                <Item
                    variant="outline"
                    className="min-h-24 border-indigo-500/45 bg-indigo-500/8 hover:bg-indigo-500/14 transition-colors duration-100"
                >
                    <ItemContent>
                        <ItemTitle className="text-2xl font-bold capitalize text-indigo-300">URL Shortener</ItemTitle>
                        <ItemDescription className="text-indigo-100/85">
                            Shorten long URLs to make them easier to share.
                        </ItemDescription>
                    </ItemContent>
                </Item>
            </Link>
            <Link to="/pastebin">
                <Item
                    variant="outline"
                    className="min-h-24 border-orange-500/45 bg-orange-500/8 hover:bg-orange-500/14 transition-colors duration-100"
                >
                    <ItemContent>
                        <ItemTitle className='text-2xl font-bold capitalize text-orange-300'>Pastebin</ItemTitle>
                        <ItemDescription className="text-orange-100/85">
                            Share text snippets with others.
                        </ItemDescription>
                    </ItemContent>
                </Item>
            </Link>
        </section>
    </main>
}