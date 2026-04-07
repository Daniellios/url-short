import { Item, ItemContent, ItemDescription, ItemTitle } from '#/components/ui/item'
import { createFileRoute, Link } from '@tanstack/react-router'


export const Route = createFileRoute('/')({ component: MainPage })

function MainPage() {
    return <main className="page-wrap px-4 pb-8 pt-14">
        <section className="island-shell rise-in max-w-2xl rounded-2xl">
            <h1 className="display-title mb-2 text-3xl sm:text-4xl">Utilitools</h1>
            <p className="mb-6 text-sm opacity-80">A collection of useful tools to help you with your daily tasks.</p>
        </section>

        <section className="grid grid-cols-2 gap-4">
            <Link to="/shorten-url">
                <Item variant="outline" className='hover:bg-muted/10 transition-colors duration-100 min-h-24'>
                    <ItemContent>
                        <ItemTitle className='text-xl'>URL Shortener</ItemTitle>
                        <ItemDescription>
                            Shorten long URLs to make them easier to share.
                        </ItemDescription>
                    </ItemContent>
                </Item>
            </Link>
            <Link to="/pastebin">
                <Item variant="outline" className='hover:bg-muted/10 transition-colors duration-100 min-h-24'>
                    <ItemContent>
                        <ItemTitle className='text-xl'>Pastebin</ItemTitle>
                        <ItemDescription>
                            Share text snippets with others.
                        </ItemDescription>
                    </ItemContent>
                </Item>
            </Link>
        </section>
    </main>
}